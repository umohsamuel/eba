package repository

import (
	"context"
	"database/sql"

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
		return nil, domainerrors.ErrMFANotEnrolled.Wrap(err)
	}

	return &mfa.TOTP{
		UserID:          t.UserID,
		SecretEncrypted: t.SecretEncrypted,
		ConfirmedAt:     &t.ConfirmedAt.Time,
		LastUsedStep:    &t.LastUsedStep.Int64,
		CreatedAt:       t.CreatedAt,
	}, nil
}

func (m *MFARepositoryAdapter) IsEnabled(ctx context.Context, userID uuid.UUID) (bool, error) {
	enabled, err := m.queries.IsUserMFAEnabled(ctx, userID)
	if err != nil {
		return false, domainerrors.ErrMFANotEnrolled.Wrap(err)
	}

	return enabled, nil
}

func (m *MFARepositoryAdapter) SavePendingTOTP(ctx context.Context, userID uuid.UUID, secretEncrypted []byte) error {
	count, err := m.queries.UpsertPendingUserTOTP(ctx, sqlc.UpsertPendingUserTOTPParams{
		UserID:          userID,
		SecretEncrypted: secretEncrypted,
	})
	if err != nil {
		return domainerrors.ErrMFAAlreadyEnabled.Wrap(err)
	}

	if count > 0 {
		return nil
	}

	return domainerrors.ErrInternal.Wrap(err)

}

func (m *MFARepositoryAdapter) ConfirmTOTP(ctx context.Context, userID uuid.UUID, step int64, recoveryCodeHashes []string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}
	defer tx.Rollback()

	q := m.queries.WithTx(tx)

	_, err = q.ConfirmUserTOTP(ctx, sqlc.ConfirmUserTOTPParams{
		UserID:       userID,
		LastUsedStep: step,
	})
	if err != nil {
		return domainerrors.ErrMFANotEnrolled.Wrap(err)
	}

	recoveryCodeId := uuid.New()

	for i, _ := range recoveryCodeHashes {
		err = q.InsertUserRecoveryCode(ctx, sqlc.InsertUserRecoveryCodeParams{
			ID:       recoveryCodeId,
			UserID:   userID,
			CodeHash: recoveryCodeHashes[i],
		})
		if err != nil {
			return domainerrors.ErrMFACodeInvalid.Wrap(err)
		}
	}

	return tx.Commit()
}

func (m *MFARepositoryAdapter) AdvanceStep(ctx context.Context, userID uuid.UUID, step int64) error {
	count, err := m.queries.AdvanceUserTOTPStep(ctx, sqlc.AdvanceUserTOTPStepParams{
		UserID:       userID,
		LastUsedStep: step,
	})
	if err != nil {
		return domainerrors.ErrMFACodeInvalid.Wrap(err)
	}

	if count > 0 {
		return nil
	}

	return domainerrors.ErrMFACodeInvalid.Wrap(err)

}

func (m *MFARepositoryAdapter) Disable(ctx context.Context, userID uuid.UUID) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}
	defer tx.Rollback()

	q := m.queries.WithTx(tx)

	err = q.DeleteUserTOTP(ctx, userID)
	if err != nil {
		return domainerrors.ErrMFANotEnrolled.Wrap(err)
	}

	err = q.DeleteUserRecoveryCodes(ctx, userID)
	if err != nil {
		return domainerrors.ErrMFANotEnrolled.Wrap(err)
	}

	return tx.Commit()
}

func (m *MFARepositoryAdapter) ReplaceRecoveryCodes(ctx context.Context, userID uuid.UUID, codeHashes []string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return domainerrors.ErrInternal.Wrap(err)
	}
	defer tx.Rollback()

	q := m.queries.WithTx(tx)

	err = q.DeleteUserRecoveryCodes(ctx, userID)
	if err != nil {
		return domainerrors.ErrMFANotEnrolled.Wrap(err)
	}

	codeId := uuid.New()

	for i, _ := range codeHashes {
		err = q.InsertUserRecoveryCode(ctx, sqlc.InsertUserRecoveryCodeParams{
			ID:       codeId,
			UserID:   userID,
			CodeHash: codeHashes[i],
		})
		if err != nil {
			return domainerrors.ErrMFACodeInvalid.Wrap(err)
		}
	}

	return tx.Commit()

}

func (m *MFARepositoryAdapter) ConsumeRecoveryCode(ctx context.Context, userID uuid.UUID, codeHash string) error {
	count, err := m.queries.ConsumeUserRecoveryCode(ctx, sqlc.ConsumeUserRecoveryCodeParams{
		UserID:   userID,
		CodeHash: codeHash,
	})
	if err != nil {
		return domainerrors.ErrMFACodeInvalid.Wrap(err)
	}

	if count > 0 {
		return nil
	}

	return domainerrors.ErrInternal

}

func (m *MFARepositoryAdapter) CountRecoveryCodes(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := m.queries.CountUnusedUserRecoveryCodes(ctx, userID)
	if err != nil {
		return 0, domainerrors.ErrMFANotEnrolled.Wrap(err)
	}

	return count, nil

}
