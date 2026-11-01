package services

import (
	"github.com/umohsamuel/eba/services/auth/internal/domain/user"
	authService "github.com/umohsamuel/eba/services/auth/internal/services/auth"
)

type Services struct {
	Auth *authService.AuthService
}

type Config struct {
	Users user.UserRepository
}

func New(deps Config) *Services {
	return &Services{
		Auth: authService.NewAuthService(deps.Users),
	}
}
