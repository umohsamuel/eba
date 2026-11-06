package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/umohsamuel/eba/services/auth/internal/adapters/db/sqlc"
	domainerrors "github.com/umohsamuel/eba/services/auth/internal/domain/errors"
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
		ID:        _refreshToken.ID,
		UserID:    _refreshToken.UserID,
		TokenHash: _refreshToken.TokenHash,
		ExpiresAt: _refreshToken.ExpiresAt,
	})
	if err != nil {
		if isForeignKeyViolation(err) {
			return nil, domainerrors.ErrUserNotFound.Wrap(err)
		}
		return nil, domainerrors.ErrInternal.Wrap(err)
	}

	_refreshToken.Revoked = createdToken.Revoked

	return _refreshToken, nil
}

func (r *RefreshRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*refreshtoken.RefreshToken, error) {
	token, err := r.queries.GetRefreshToken(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrTokenInvalid.Wrap(err)
		}
		return nil, domainerrors.ErrInternal.Wrap(err)
	}

	return &refreshtoken.RefreshToken{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		Revoked:   token.Revoked,
		ExpiresAt: token.ExpiresAt,
	}, nil
}

func (r *RefreshRepositoryAdapter) Revoke(ctx context.Context, id uuid.UUID) (*uuid.UUID, error) {
	revokedToken, err := r.queries.RevokeRefreshToken(ctx, sqlc.RevokeRefreshTokenParams{
		ID:      id,
		Revoked: true,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrRefreshTokenNotFound.Wrap(err)
		}
		return nil, domainerrors.ErrInternal.Wrap(err)
	}

	return &revokedToken.ID, nil

}
