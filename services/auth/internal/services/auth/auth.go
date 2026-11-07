package auth

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/google/uuid"
	"github.com/umohsamuel/eba/services/auth/internal/domain/errors"
	"github.com/umohsamuel/eba/services/auth/internal/domain/refreshtoken"
	"github.com/umohsamuel/eba/services/auth/internal/domain/user"
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
}

type Config struct {
	Users        user.UserRepository
	RefreshToken refreshtoken.RefreshTokenRepository
	Env          env.Env
	TokenManager token.TokenManager
}

func NewAuthService(deps Config) *AuthService {
	return &AuthService{
		users:        deps.Users,
		refreshToken: deps.RefreshToken,
		env:          deps.Env,
		tokenManager: deps.TokenManager,
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
const dummyHash = "$2a$10$xK9vG/5F7y8Kz7v9B6h2u.D8r1e2f3g4h5i6j7k8l9m0n1o2p3q4r"

type TokenConfig struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
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
		targetHash = dummyHash
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
