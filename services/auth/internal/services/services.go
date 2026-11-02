package services

import (
	"github.com/umohsamuel/eba/services/auth/internal/adapters/db/repository"
	authService "github.com/umohsamuel/eba/services/auth/internal/services/auth"
	"github.com/umohsamuel/eba/services/auth/internal/services/refresh"
	"github.com/umohsamuel/eba/services/auth/pkg/env"
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
	return &Services{
		Auth: authService.NewAuthService(authService.Config{
			Users: deps.Repository.Users,
			Env:   deps.Env,
		}),
		Refresh: refresh.NewRefreshService(refresh.Config{
			RefreshToken: deps.Repository.Refresh,
			Env:          deps.Env,
		}),
	}
}
