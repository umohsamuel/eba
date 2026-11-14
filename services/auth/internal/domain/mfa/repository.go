package mfa

import (
	"context"

	"github.com/google/uuid"
)

type MFARepository interface {
	GetTOTP(ctx context.Context, userID uuid.UUID) (*TOTP, error)
	IsEnabled(ctx context.Context, userID uuid.UUID) (bool, error)
	SavePendingTOTP(ctx context.Context, userID uuid.UUID, secretEncrypted []byte) error
	ConfirmTOTP(ctx context.Context, userID uuid.UUID, step int64, recoveryCodeHashes []string) error
	AdvanceStep(ctx context.Context, userID uuid.UUID, step int64) error
	Disable(ctx context.Context, userID uuid.UUID) error
	ReplaceRecoveryCodes(ctx context.Context, userID uuid.UUID, codeHashes []string) error
	ConsumeRecoveryCode(ctx context.Context, userID uuid.UUID, codeHash string) error
	CountRecoveryCodes(ctx context.Context, userID uuid.UUID) (int64, error)
}
