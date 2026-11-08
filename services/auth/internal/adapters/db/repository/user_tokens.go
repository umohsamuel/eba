package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/umohsamuel/eba/services/auth/internal/adapters/db/sqlc"
	domainerrors "github.com/umohsamuel/eba/services/auth/internal/domain/errors"
	"github.com/umohsamuel/eba/services/auth/internal/domain/usertokens"
)

type UserTokensRepositoryAdapter struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewUserTokensRepositoryAdapter(db *sql.DB, queries *sqlc.Queries) *UserTokensRepositoryAdapter {
	return &UserTokensRepositoryAdapter{
		db:      db,
		queries: queries,
	}
}

var _ usertokens.UserTokensRepository = (*UserTokensRepositoryAdapter)(nil)

func (u *UserTokensRepositoryAdapter) Create(ctx context.Context, token *usertokens.UserToken) (*usertokens.UserToken, error) {
	ut, err := u.queries.CreateUserToken(ctx, sqlc.CreateUserTokenParams{
		ID:        token.ID,
		UserID:    token.UserID,
		Purpose:   token.Purpose,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
	})
	if err != nil {
		if isForeignKeyViolation(err) {
			return nil, domainerrors.ErrUserNotFound.Wrap(err)
		}
		return nil, domainerrors.ErrInternal.Wrap(err)
	}

	return &usertokens.UserToken{
		ID:        ut.ID,
		UserID:    ut.UserID,
		Purpose:   ut.Purpose,
		TokenHash: ut.TokenHash,
		ExpiresAt: ut.ExpiresAt,
		UsedAt:    usedAtPtr(ut.UsedAt),
	}, nil
}

func (u *UserTokensRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*usertokens.UserToken, error) {
	ut, err := u.queries.GetUserTokenByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrNotFound.Wrap(err)
		}
		return nil, domainerrors.ErrInternal.Wrap(err)
	}

	return &usertokens.UserToken{
		ID:        ut.ID,
		UserID:    ut.UserID,
		Purpose:   ut.Purpose,
		TokenHash: ut.TokenHash,
		ExpiresAt: ut.ExpiresAt,
		UsedAt:    usedAtPtr(ut.UsedAt),
	}, nil
}

func (u *UserTokensRepositoryAdapter) GetByHash(ctx context.Context, hash string) (*usertokens.UserToken, error) {
	ut, err := u.queries.GetUserTokenByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrNotFound.Wrap(err)
		}
		return nil, domainerrors.ErrInternal.Wrap(err)
	}

	return &usertokens.UserToken{
		ID:        ut.ID,
		UserID:    ut.UserID,
		Purpose:   ut.Purpose,
		TokenHash: ut.TokenHash,
		ExpiresAt: ut.ExpiresAt,
		UsedAt:    usedAtPtr(ut.UsedAt),
	}, nil
}

func (u *UserTokensRepositoryAdapter) List(ctx context.Context, filter usertokens.UserTokensFilter) ([]*usertokens.UserToken, error) {
	ut, err := u.queries.ListUserTokens(ctx)
	if err != nil {
		return nil, domainerrors.ErrInternal.Wrap(err)
	}

	usertokenMap := make([]*usertokens.UserToken, 0, len(ut))

	for _, t := range ut {
		usertokenMap = append(usertokenMap, &usertokens.UserToken{
			ID:        t.ID,
			UserID:    t.UserID,
			Purpose:   t.Purpose,
			TokenHash: t.TokenHash,
			ExpiresAt: t.ExpiresAt,
			UsedAt:    usedAtPtr(t.UsedAt),
		})
	}

	return usertokenMap, nil
}

func (u *UserTokensRepositoryAdapter) Revoke(ctx context.Context, id uuid.UUID) (*usertokens.UserToken, error) {
	ut, err := u.queries.RevokeUserToken(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrTokenInvalid.Wrap(err)
		}
		return nil, domainerrors.ErrInternal.Wrap(err)
	}

	return &usertokens.UserToken{
		ID:        ut.ID,
		UserID:    ut.UserID,
		Purpose:   ut.Purpose,
		TokenHash: ut.TokenHash,
		ExpiresAt: ut.ExpiresAt,
		UsedAt:    usedAtPtr(ut.UsedAt),
	}, nil
}

func (u *UserTokensRepositoryAdapter) Delete(ctx context.Context, id uuid.UUID) (*uuid.UUID, error) {
	tokenId, err := u.queries.DeleteUserToken(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrNotFound.Wrap(err)
		}
		return nil, domainerrors.ErrInternal.Wrap(err)
	}

	return &tokenId, nil
}

func (u *UserTokensRepositoryAdapter) DeleteAllStale(ctx context.Context) (int64, error) {
	count, err := u.queries.DeleteStaleUserTokens(ctx)
	if err != nil {
		return 0, domainerrors.ErrInternal.Wrap(err)
	}

	return count, nil
}

func usedAtPtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func (u *UserTokensRepositoryAdapter) RevokeAllForUser(ctx context.Context, userID uuid.UUID, purpose string) error {
	err := u.queries.RevokeUserTokensForUser(ctx, sqlc.RevokeUserTokensForUserParams{
		UserID:  userID,
		Purpose: purpose,
	})
	if err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}

	return nil
}

func (u *UserTokensRepositoryAdapter) ConsumePasswordReset(ctx context.Context, tokenHash, passwordHash string) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}
	defer tx.Rollback()

	q := u.queries.WithTx(tx)

	userID, err := q.ConsumeUserToken(ctx, sqlc.ConsumeUserTokenParams{
		TokenHash: tokenHash,
		Purpose:   usertokens.PurposePasswordReset,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainerrors.ErrTokenInvalid.Wrap(err)
		}
		return domainerrors.ErrInternal.Wrap(err)
	}

	updated, err := q.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}
	if updated == 0 {
		return domainerrors.ErrUserNotFound
	}

	if err := q.RevokeRefreshTokensForUser(ctx, userID); err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}

	if err := tx.Commit(); err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}

	return nil
}
