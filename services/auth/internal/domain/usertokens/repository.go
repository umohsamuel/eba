package usertokens

import (
	"context"

	"github.com/google/uuid"
)

type UserTokensFilter struct {
	Limit  int
	Offset int
}

type UserTokensRepository interface {
	Create(ctx context.Context, token *UserToken) (*UserToken, error)
	GetByID(ctx context.Context, id uuid.UUID) (*UserToken, error)
	GetByHash(ctx context.Context, hash string) (*UserToken, error)
	List(ctx context.Context, filter UserTokensFilter) ([]*UserToken, error)
	Revoke(ctx context.Context, id uuid.UUID) (*UserToken, error)
	RevokeAllForUser(ctx context.Context, userID uuid.UUID, purpose string) error
	ConsumePasswordReset(ctx context.Context, tokenHash, passwordHash string) error
	Delete(ctx context.Context, id uuid.UUID) (*uuid.UUID, error)
	DeleteAllStale(ctx context.Context) (int64, error)
}
