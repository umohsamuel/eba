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

	// The repository returns domain errors: ErrEmailAlreadyExists if a
	// concurrent request registered the same email first, ErrInternal otherwise.
	createdUser, err := s.users.Create(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

// A dummy hash to run against if a user isn't found to mitigate timing attacks.
const dummyHash = "$2a$10$xK9vG/5F7y8Kz7v9B6h2u.D8r1e2f3g4h5i6j7k8l9m0n1o2p3q4r"

type TokenConfig struct {
	AccessToken  string
	RefreshToken string
}

type LoginResponse struct {
	User  *user.User
	Token *TokenConfig
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

	accessToken, err := s.tokenManager.GenerateAccessToken(*u, 15*time.Minute)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}

	id := uuid.New()
	issuedAt := time.Now().UTC().Truncate(time.Second)
	expiresAt := issuedAt.Add(7 * 24 * time.Hour)

	refershToken, err := s.tokenManager.GenerateRefreshToken(*u, id.String(), issuedAt, expiresAt)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}

	createdRefreshToken, err := s.refreshToken.Create(ctx, &refreshtoken.RefreshToken{
		ID:        id,
		UserID:    u.ID,
		Token:     refershToken,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}

	return &LoginResponse{
		User: u,
		Token: &TokenConfig{
			AccessToken:  accessToken,
			RefreshToken: createdRefreshToken.Token,
		},
	}, nil
}
