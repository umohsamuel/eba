package auth

import (
	"context"
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
}

type Config struct {
	Users        user.UserRepository
	RefreshToken refreshtoken.RefreshTokenRepository
	Env          env.Env
}

func NewAuthService(deps Config) *AuthService {
	return &AuthService{
		users:        deps.Users,
		refreshToken: deps.RefreshToken,
		env:          deps.Env,
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
		return nil, errors.ErrInternal
	}

	_, err = s.users.GetByEmail(ctx, email)
	if err == nil {
		return nil, errors.ErrEmailAlreadyExists
	}

	newUser := &user.User{
		Email:        email,
		PasswordHash: passwordHash,
	}

	createdUser, err := s.users.Create(ctx, newUser)
	if err != nil {
		return nil, errors.ErrInternal
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

	accessToken, err := token.GenerateAccessToken(*u, []byte(s.env.JWT.JWT_SECRET), 15*time.Minute)

	id := uuid.New()
	refershToken, err := token.GenerateRefreshToken(*u, []byte(s.env.JWT.JWT_REFRESH_SECRET), 7*24*time.Hour)

	createdRefreshToken, err := s.refreshToken.Create(ctx, &refreshtoken.RefreshToken{
		UserID: u.ID,
		Token:  refershToken,
	})
	if err != nil {
		return nil, errors.ErrInternal
	}

	return &LoginResponse{
		User: u,
		Token: &TokenConfig{
			AccessToken:  accessToken,
			RefreshToken: createdRefreshToken.Token,
		},
	}, nil
}
