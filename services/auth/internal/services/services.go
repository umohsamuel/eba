package services

import (
	"github.com/umohsamuel/eba/services/auth/internal/adapters/db/repository"
	authService "github.com/umohsamuel/eba/services/auth/internal/services/auth"
	mfaService "github.com/umohsamuel/eba/services/auth/internal/services/mfa"
	"github.com/umohsamuel/eba/services/auth/internal/services/refresh"
	"github.com/umohsamuel/eba/services/auth/pkg/env"
	"github.com/umohsamuel/eba/services/auth/pkg/secretbox"
	"github.com/umohsamuel/eba/services/auth/pkg/token"
)

type Services struct {
	Auth    *authService.AuthService
	Refresh *refresh.RefreshService
	MFA     *mfaService.Service
}

type Config struct {
	Repository *repository.AdaptersRepository
	Env        env.Env
	AccessKeys *token.AccessKeys
	SecretBox  secretbox.SecretBox
}

func New(deps Config) *Services {
	tokenManager := token.NewTokenManager(
		deps.Env.JWT.JWT_ISSUER,
		deps.AccessKeys,
		deps.Env.JWT.JWT_AUDIENCE,
		[]byte(deps.Env.JWT.JWT_REFRESH_SECRET),
		[]string{deps.Env.JWT.JWT_REFRESH_AUDIENCE},
	)

	mfaSvc := mfaService.New(mfaService.Config{
		Repository: deps.Repository.MFA,
		Users:      deps.Repository.Users,
		SecretBox:  deps.SecretBox,
		Issuer:     deps.Env.MFA_ISSUER,
	})

	return &Services{
		MFA: mfaSvc,
		Auth: authService.NewAuthService(authService.Config{
			Users:        deps.Repository.Users,
			RefreshToken: deps.Repository.Refresh,
			Env:          deps.Env,
			TokenManager: *tokenManager,
			UserToken:    deps.Repository.UserToken,
			MFA:          deps.Repository.MFA,
			MFAVerifier:  mfaSvc,
		}),
		Refresh: refresh.NewRefreshService(refresh.Config{
			RefreshToken: deps.Repository.Refresh,
			Env:          deps.Env,
		}),
	}
}
