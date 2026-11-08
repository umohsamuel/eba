package usertokens

import (
	"time"

	"github.com/google/uuid"
)

const PurposePasswordReset = "password_reset"

type UserToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Purpose   string     `json:"purpose"`
	TokenHash string     `json:"token_hash"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
}
