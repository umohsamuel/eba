package refresh

import (
	"context"

	"github.com/google/uuid"
	domainErrors "github.com/umohsamuel/eba/services/auth/internal/domain/errors"
	"github.com/umohsamuel/eba/services/auth/internal/domain/refreshtoken"
	"github.com/umohsamuel/eba/services/auth/pkg/env"
)

type RefreshService struct {
	refreshToken refreshtoken.RefreshTokenRepository
	env          env.Env
}

type Config struct {
	RefreshToken refreshtoken.RefreshTokenRepository
	Env          env.Env
}

func NewRefreshService(deps Config) *RefreshService {
	return &RefreshService{
		refreshToken: deps.RefreshToken,
		env:          deps.Env,
	}
}

func (s *RefreshService) Create(ctx context.Context, _refreshToken *refreshtoken.RefreshToken) (*refreshtoken.RefreshToken, error) {
	if _refreshToken == nil || _refreshToken.UserID == uuid.Nil || _refreshToken.Token == "" {
		return nil, domainErrors.ErrBadRequest
	}

	createdToken, err := s.refreshToken.Create(ctx, &refreshtoken.RefreshToken{
		UserID: _refreshToken.UserID,
		Token:  _refreshToken.Token,
	})
	if err != nil {
		return nil, domainErrors.ErrInternal
	}

	return createdToken, nil
}

func (s *RefreshService) GetByID(ctx context.Context, id uuid.UUID) (*refreshtoken.RefreshToken, error) {
	if id == uuid.Nil {
		return nil, domainErrors.ErrBadRequest
	}

	token, err := s.refreshToken.GetByID(ctx, id)
	if err != nil {
		return nil, domainErrors.ErrInternal
	}

	return token, nil
}

func (s *RefreshService) Revoke(ctx context.Context, id uuid.UUID) (*uuid.UUID, error) {
	if id == uuid.Nil {
		return nil, domainErrors.ErrBadRequest
	}

	revokedToken, err := s.refreshToken.Revoke(ctx, id)
	if err != nil {
		return nil, domainErrors.ErrInternal
	}

	return revokedToken, nil
}
