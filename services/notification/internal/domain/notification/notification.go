package notification

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusSent    Status = "sent"
	StatusFailed  Status = "failed"
)

type Notification struct {
	ID                uuid.UUID
	EventID           string
	Channel           string
	Template          string
	Recipient         string
	Status            Status
	Attempts          int
	LastError         string
	ProviderMessageID string
	CreatedAt         time.Time
	SentAt            *time.Time
}

var (
	ErrInvalidRequest  = errors.New("invalid notification request")
	ErrUnknownTemplate = errors.New("unknown notification template")
	ErrUnknownChannel  = errors.New("unknown notification channel")

	ErrRecipientRejected = errors.New("recipient rejected")
)

func IsPermanent(err error) bool {
	return errors.Is(err, ErrInvalidRequest) ||
		errors.Is(err, ErrUnknownTemplate) ||
		errors.Is(err, ErrUnknownChannel) ||
		errors.Is(err, ErrRecipientRejected)
}
