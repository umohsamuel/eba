package services

import (
	"github.com/umohsamuel/eba/services/auth/internal/domain/user"
	authService "github.com/umohsamuel/eba/services/auth/internal/services/auth"
	"github.com/umohsamuel/eba/services/auth/pkg/env"
)

type Services struct {
	Auth *authService.AuthService
}

type Config struct {
	Users user.UserRepository
	Env   env.Env
}

func New(deps Config) *Services {
	return &Services{
		Auth: authService.NewAuthService(authService.Config{
			Users: deps.Users,
			Env:   deps.Env,
		}),
	}
}
