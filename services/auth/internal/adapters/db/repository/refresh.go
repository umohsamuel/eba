package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/umohsamuel/eba/services/auth/internal/adapters/db/sqlc"
	"github.com/umohsamuel/eba/services/auth/internal/domain/refreshtoken"
)

type RefreshRepositoryAdapter struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewRefreshRepositoryAdapter(db *sql.DB, queries *sqlc.Queries) *RefreshRepositoryAdapter {
	return &RefreshRepositoryAdapter{
		db:      db,
		queries: queries,
	}
}

var _ refreshtoken.RefreshTokenRepository = (*RefreshRepositoryAdapter)(nil)

func (r *RefreshRepositoryAdapter) Create(ctx context.Context, _refreshToken *refreshtoken.RefreshToken) (*refreshtoken.RefreshToken, error) {
	createdToken, err := r.queries.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		UserID: _refreshToken.UserID,
		Token:  _refreshToken.Token,
	})
	if err != nil {
		return nil, err
	}

	_refreshToken.ID = createdToken.ID
	_refreshToken.Revoked = createdToken.Revoked

	return _refreshToken, nil
}

func (r *RefreshRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*refreshtoken.RefreshToken, error) {
	token, err := r.queries.GetRefreshToken(ctx, id)
	if err != nil {
		return nil, err
	}

	return &refreshtoken.RefreshToken{
		ID:      token.ID,
		UserID:  token.UserID,
		Token:   token.Token,
		Revoked: token.Revoked,
	}, nil
}

func (r *RefreshRepositoryAdapter) Revoke(ctx context.Context, id uuid.UUID) (*uuid.UUID, error) {
	revokedToken, err := r.queries.RevokeRefreshToken(ctx, sqlc.RevokeRefreshTokenParams{
		ID:      id,
		Revoked: true,
	})
	if err != nil {
		return nil, err
	}

	return &revokedToken.ID, err

}
