package notification

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateIfAbsent(ctx context.Context, n *Notification) (*Notification, error)
	MarkSent(ctx context.Context, id uuid.UUID, providerMessageID string) error
	MarkFailed(ctx context.Context, id uuid.UUID, lastError string) error
}
