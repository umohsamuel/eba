package refreshtoken

import (
	"context"

	"github.com/google/uuid"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, refreshToken *RefreshToken) (*RefreshToken, error)
	GetByID(ctx context.Context, id uuid.UUID) (*RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) (*uuid.UUID, error)
}
