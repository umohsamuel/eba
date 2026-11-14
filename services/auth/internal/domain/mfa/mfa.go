package mfa

import (
	"time"

	"github.com/google/uuid"
)

type TOTP struct {
	UserID          uuid.UUID
	SecretEncrypted []byte
	ConfirmedAt     *time.Time
	LastUsedStep    *int64
	CreatedAt       time.Time
}

func (t *TOTP) Enabled() bool {
	return t.ConfirmedAt != nil
}
