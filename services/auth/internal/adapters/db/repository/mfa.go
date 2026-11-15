package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/umohsamuel/eba/services/auth/internal/adapters/db/sqlc"
	domainerrors "github.com/umohsamuel/eba/services/auth/internal/domain/errors"
	"github.com/umohsamuel/eba/services/auth/internal/domain/mfa"
)

type MFARepositoryAdapter struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewMFARepository(db *sql.DB, queries *sqlc.Queries) *MFARepositoryAdapter {
	return &MFARepositoryAdapter{
		db:      db,
		queries: queries,
	}
}

var _ mfa.MFARepository = (*MFARepositoryAdapter)(nil)

func (m *MFARepositoryAdapter) GetTOTP(ctx context.Context, userID uuid.UUID) (*mfa.TOTP, error) {
	t, err := m.queries.GetUserTOTP(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrMFANotEnrolled.Wrap(err)
		}
		return nil, domainerrors.ErrInternal.Wrap(err)
	}

	return &mfa.TOTP{
		UserID:          t.UserID,
		SecretEncrypted: t.SecretEncrypted,
		ConfirmedAt:     usedAtPtr(t.ConfirmedAt),
		LastUsedStep:    int64Ptr(t.LastUsedStep),
		CreatedAt:       t.CreatedAt,
	}, nil
}

func (m *MFARepositoryAdapter) IsEnabled(ctx context.Context, userID uuid.UUID) (bool, error) {
	enabled, err := m.queries.IsUserMFAEnabled(ctx, userID)
	if err != nil {
		return false, domainerrors.ErrInternal.Wrap(err)
	}

	return enabled, nil
}

func (m *MFARepositoryAdapter) SavePendingTOTP(ctx context.Context, userID uuid.UUID, secretEncrypted []byte) error {
	count, err := m.queries.UpsertPendingUserTOTP(ctx, sqlc.UpsertPendingUserTOTPParams{
		UserID:          userID,
		SecretEncrypted: secretEncrypted,
	})
	if err != nil {
		if isForeignKeyViolation(err) {
			return domainerrors.ErrUserNotFound.Wrap(err)
		}
		return domainerrors.ErrInternal.Wrap(err)
	}

	if count == 0 {
		return domainerrors.ErrMFAAlreadyEnabled
	}

	return nil
}

func (m *MFARepositoryAdapter) ConfirmTOTP(ctx context.Context, userID uuid.UUID, step int64, recoveryCodeHashes []string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}
	defer tx.Rollback()

	q := m.queries.WithTx(tx)

	count, err := q.ConfirmUserTOTP(ctx, sqlc.ConfirmUserTOTPParams{
		UserID:       userID,
		LastUsedStep: sql.NullInt64{Int64: step, Valid: true},
	})
	if err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}
	if count == 0 {
		return domainerrors.ErrMFANotEnrolled
	}

	if err := insertRecoveryCodes(ctx, q, userID, recoveryCodeHashes); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}

	return nil
}

func (m *MFARepositoryAdapter) AdvanceStep(ctx context.Context, userID uuid.UUID, step int64) error {
	count, err := m.queries.AdvanceUserTOTPStep(ctx, sqlc.AdvanceUserTOTPStepParams{
		UserID:       userID,
		LastUsedStep: sql.NullInt64{Int64: step, Valid: true},
	})
	if err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}

	if count == 0 {
		return domainerrors.ErrMFACodeInvalid
	}

	return nil
}

func (m *MFARepositoryAdapter) Disable(ctx context.Context, userID uuid.UUID) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}
	defer tx.Rollback()

	q := m.queries.WithTx(tx)

	if err := q.DeleteUserTOTP(ctx, userID); err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}

	if err := q.DeleteUserRecoveryCodes(ctx, userID); err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}

	if err := tx.Commit(); err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}

	return nil
}

func (m *MFARepositoryAdapter) ReplaceRecoveryCodes(ctx context.Context, userID uuid.UUID, codeHashes []string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}
	defer tx.Rollback()

	q := m.queries.WithTx(tx)

	if err := q.DeleteUserRecoveryCodes(ctx, userID); err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}

	if err := insertRecoveryCodes(ctx, q, userID, codeHashes); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}

	return nil
}

func (m *MFARepositoryAdapter) ConsumeRecoveryCode(ctx context.Context, userID uuid.UUID, codeHash string) error {
	count, err := m.queries.ConsumeUserRecoveryCode(ctx, sqlc.ConsumeUserRecoveryCodeParams{
		UserID:   userID,
		CodeHash: codeHash,
	})
	if err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}

	if count == 0 {
		return domainerrors.ErrMFACodeInvalid
	}

	return nil
}

func (m *MFARepositoryAdapter) CountRecoveryCodes(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := m.queries.CountUnusedUserRecoveryCodes(ctx, userID)
	if err != nil {
		return 0, domainerrors.ErrInternal.Wrap(err)
	}

	return count, nil
}

func insertRecoveryCodes(ctx context.Context, q *sqlc.Queries, userID uuid.UUID, codeHashes []string) error {
	for _, codeHash := range codeHashes {
		err := q.InsertUserRecoveryCode(ctx, sqlc.InsertUserRecoveryCodeParams{
			ID:       uuid.New(),
			UserID:   userID,
			CodeHash: codeHash,
		})
		if err != nil {
			if isForeignKeyViolation(err) {
				return domainerrors.ErrUserNotFound.Wrap(err)
			}
			return domainerrors.ErrInternal.Wrap(err)
		}
	}

	return nil
}

func int64Ptr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	return &n.Int64
}
