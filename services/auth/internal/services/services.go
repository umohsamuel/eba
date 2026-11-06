package services

import (
	"github.com/umohsamuel/eba/services/auth/internal/adapters/db/repository"
	authService "github.com/umohsamuel/eba/services/auth/internal/services/auth"
	"github.com/umohsamuel/eba/services/auth/internal/services/refresh"
	"github.com/umohsamuel/eba/services/auth/pkg/env"
	"github.com/umohsamuel/eba/services/auth/pkg/token"
)

type Services struct {
	Auth    *authService.AuthService
	Refresh *refresh.RefreshService
}

type Config struct {
	Repository *repository.AdaptersRepository
	Env        env.Env
}

func New(deps Config) *Services {
	tokenManager := token.NewTokenManager(
		deps.Env.JWT.JWT_ISSUER,
		[]byte(deps.Env.JWT.JWT_SECRET),
		deps.Env.JWT.JWT_AUDIENCE,
		[]byte(deps.Env.JWT.JWT_REFRESH_SECRET),
		deps.Env.JWT.JWT_AUDIENCE,
	)

	return &Services{
		Auth: authService.NewAuthService(authService.Config{
			Users:        deps.Repository.Users,
			RefreshToken: deps.Repository.Refresh,
			Env:          deps.Env,
			TokenManager: *tokenManager,
		}),
		Refresh: refresh.NewRefreshService(refresh.Config{
			RefreshToken: deps.Repository.Refresh,
			Env:          deps.Env,
		}),
	}
}
