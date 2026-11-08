package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	stderrors "errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/umohsamuel/eba/services/auth/internal/domain/errors"
	"github.com/umohsamuel/eba/services/auth/internal/domain/refreshtoken"
	"github.com/umohsamuel/eba/services/auth/internal/domain/user"
	"github.com/umohsamuel/eba/services/auth/internal/domain/usertokens"
	"github.com/umohsamuel/eba/services/auth/pkg/env"
	"github.com/umohsamuel/eba/services/auth/pkg/hash"
	"github.com/umohsamuel/eba/services/auth/pkg/regexx"
	"github.com/umohsamuel/eba/services/auth/pkg/token"
)

type AuthService struct {
	users        user.UserRepository
	refreshToken refreshtoken.RefreshTokenRepository
	env          env.Env
	tokenManager token.TokenManager
	usertoken    usertokens.UserTokensRepository
	notifier     PasswordResetNotifier
}

type PasswordResetNotifier interface {
	SendPasswordReset(ctx context.Context, email, token string) error
}

type Config struct {
	Users        user.UserRepository
	RefreshToken refreshtoken.RefreshTokenRepository
	Env          env.Env
	TokenManager token.TokenManager
	UserToken    usertokens.UserTokensRepository
	Notifier     PasswordResetNotifier
}

func NewAuthService(deps Config) *AuthService {
	dummyHash()

	return &AuthService{
		users:        deps.Users,
		refreshToken: deps.RefreshToken,
		env:          deps.Env,
		tokenManager: deps.TokenManager,
		usertoken:    deps.UserToken,
		notifier:     deps.Notifier,
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	email string,
	password string,
) (*user.User, error) {
	if !regexx.ValidateEmail(email) {
		return nil, errors.ErrInvalidEmail
	}

	if !regexx.ValidatePassword(password) {
		return nil, errors.ErrInvalidPassword
	}

	passwordHash, err := hash.HashPassword(password)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}

	_, err = s.users.GetByEmail(ctx, email)
	switch {
	case err == nil:
		return nil, errors.ErrEmailAlreadyExists
	case !stderrors.Is(err, errors.ErrUserNotFound):
		return nil, err
	}

	newUser := &user.User{
		Email:        email,
		PasswordHash: passwordHash,
	}

	createdUser, err := s.users.Create(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

// A dummy hash to run against if a user isn't found to mitigate timing attacks.
var dummyHash = sync.OnceValue(func() string {
	h, err := hash.HashPassword(uuid.NewString())
	if err != nil {
		panic(fmt.Sprintf("generate dummy password hash: %v", err))
	}
	return h
})

type TokenConfig struct {
	AccessToken           string    `json:"access_token,omitempty"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token,omitempty"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

type LoginResponse struct {
	User  *user.User   `json:"user"`
	Token *TokenConfig `json:"token"`
}

func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (*LoginResponse, error) {
	if !regexx.ValidateEmail(email) {
		return nil, errors.ErrInvalidEmail
	}

	if !regexx.ValidatePassword(password) {
		return nil, errors.ErrInvalidPassword
	}

	u, err := s.users.GetByEmail(ctx, email)
	if err != nil && !stderrors.Is(err, errors.ErrUserNotFound) {
		return nil, err
	}

	var targetHash string
	if err != nil {
		targetHash = dummyHash()
	} else {
		targetHash = u.PasswordHash
	}

	isCorrectPassword := hash.CheckPasswordHash(password, targetHash)

	if err != nil || !isCorrectPassword {
		return nil, errors.ErrInvalidCredentials
	}

	// JWT timestamps have second precision; truncating keeps the returned
	// expiry times identical to the exp claims.
	issuedAt := time.Now().UTC().Truncate(time.Second)
	accessExpiresAt := issuedAt.Add(s.env.JWT.JWT_ACCESS_TTL)
	expiresAt := issuedAt.Add(s.env.JWT.JWT_REFRESH_TTL)

	accessToken, err := s.tokenManager.GenerateAccessToken(*u, issuedAt, accessExpiresAt)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}

	id := uuid.New()

	rawRefreshToken, err := s.tokenManager.GenerateRefreshToken(*u, id.String(), issuedAt, expiresAt)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}

	_, err = s.refreshToken.Create(ctx, &refreshtoken.RefreshToken{
		ID:        id,
		UserID:    u.ID,
		TokenHash: hash.HashToken(rawRefreshToken),
		ExpiresAt: expiresAt,
		Revoked:   false,
	})
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		User: u,
		Token: &TokenConfig{
			AccessToken:           accessToken,
			AccessTokenExpiresAt:  accessExpiresAt,
			RefreshToken:          rawRefreshToken,
			RefreshTokenExpiresAt: expiresAt,
		},
	}, nil
}

func (s *AuthService) ForgotPassword(
	ctx context.Context,
	email string,
) error {
	if !regexx.ValidateEmail(email) {
		return errors.ErrInvalidEmail
	}

	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if stderrors.Is(err, errors.ErrUserNotFound) {
			return nil
		}
		return err
	}

	rawToken, err := newOpaqueToken()
	if err != nil {
		return errors.ErrInternal.Wrap(err)
	}

	if err := s.usertoken.RevokeAllForUser(ctx, u.ID, usertokens.PurposePasswordReset); err != nil {
		return err
	}

	_, err = s.usertoken.Create(ctx, &usertokens.UserToken{
		ID:        uuid.New(),
		UserID:    u.ID,
		Purpose:   usertokens.PurposePasswordReset,
		TokenHash: hash.HashToken(rawToken),
		ExpiresAt: time.Now().Add(s.env.PASSWORD_RESET_TTL),
	})
	if err != nil {
		return err
	}

	if s.notifier == nil {
		slog.ErrorContext(ctx, "password reset notifier not configured", "user_id", u.ID)
		return nil
	}

	if err := s.notifier.SendPasswordReset(ctx, u.Email, rawToken); err != nil {
		slog.ErrorContext(ctx, "send password reset", "user_id", u.ID, "error", err)
	}

	return nil
}

func (s *AuthService) ResetPassword(
	ctx context.Context,
	rawToken string,
	newPassword string,
) error {
	if rawToken == "" {
		return errors.ErrTokenInvalid
	}

	if !regexx.ValidatePassword(newPassword) {
		return errors.ErrInvalidPassword
	}

	tokenHash := hash.HashToken(rawToken)

	ut, err := s.usertoken.GetByHash(ctx, tokenHash)
	if err != nil {
		if stderrors.Is(err, errors.ErrNotFound) {
			return errors.ErrTokenInvalid.Wrap(err)
		}
		return err
	}

	if ut.Purpose != usertokens.PurposePasswordReset || ut.UsedAt != nil || time.Now().After(ut.ExpiresAt) {
		return errors.ErrTokenInvalid
	}

	passwordHash, err := hash.HashPassword(newPassword)
	if err != nil {
		return errors.ErrInternal.Wrap(err)
	}

	return s.usertoken.ConsumePasswordReset(ctx, tokenHash, passwordHash)
}

func newOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *AuthService) RefreshAccessToken(
	ctx context.Context,
	refreshToken string,
) (*TokenConfig, error) {
	claims, err := s.tokenManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		if stderrors.Is(err, token.ErrTokenExpired) {
			return nil, errors.ErrTokenExpired.Wrap(err)
		}
		return nil, errors.ErrTokenInvalid.Wrap(err)
	}

	id, err := uuid.Parse(claims.ID)
	if err != nil {
		return nil, errors.ErrTokenInvalid.Wrap(err)
	}

	row, err := s.refreshToken.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if row.Revoked || time.Now().After(row.ExpiresAt) {
		return nil, errors.ErrTokenInvalid
	}

	if isValid := hash.VerifyToken(refreshToken, row.TokenHash); !isValid {
		return nil, errors.ErrTokenInvalid
	}

	u, err := s.users.GetByID(ctx, row.UserID)
	if err != nil {
		if stderrors.Is(err, errors.ErrUserNotFound) {
			return nil, errors.ErrTokenInvalid.Wrap(err)
		}
		return nil, err
	}

	issuedAt := time.Now().UTC().Truncate(time.Second)
	accessExpiresAt := issuedAt.Add(s.env.JWT.JWT_ACCESS_TTL)

	accessToken, err := s.tokenManager.GenerateAccessToken(*u, issuedAt, accessExpiresAt)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}

	return &TokenConfig{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: row.ExpiresAt,
	}, nil
}

func (s *AuthService) JWKS() token.JWKS {
	return s.tokenManager.AccessJWKS()
}

func (s *AuthService) VerifyAccessToken(ctx context.Context, accessToken string) (*token.CustomClaims, error) {
	claims, err := s.tokenManager.ValidateAccessToken(accessToken)
	if err != nil {
		if stderrors.Is(err, token.ErrTokenExpired) {
			return nil, errors.ErrTokenExpired.Wrap(err)
		}
		return nil, errors.ErrTokenInvalid.Wrap(err)
	}

	return claims, nil
}
