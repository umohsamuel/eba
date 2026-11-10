package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/umohsamuel/eba/services/notification/internal/adapters/db/sqlc"
	domain "github.com/umohsamuel/eba/services/notification/internal/domain/notification"
)

const maxErrorLength = 1000

type NotificationRepository struct {
	queries *sqlc.Queries
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{queries: sqlc.New(db)}
}

var _ domain.Repository = (*NotificationRepository)(nil)

func (r *NotificationRepository) CreateIfAbsent(ctx context.Context, n *domain.Notification) (*domain.Notification, error) {
	err := r.queries.InsertNotification(ctx, sqlc.InsertNotificationParams{
		ID:        n.ID,
		EventID:   n.EventID,
		Channel:   n.Channel,
		Template:  n.Template,
		Recipient: n.Recipient,
		Status:    string(n.Status),
	})
	if err != nil {
		return nil, fmt.Errorf("insert notification: %w", err)
	}

	row, err := r.queries.GetNotificationByEventID(ctx, n.EventID)
	if err != nil {
		return nil, fmt.Errorf("get notification: %w", err)
	}

	return toDomain(row), nil
}

func (r *NotificationRepository) MarkSent(ctx context.Context, id uuid.UUID, providerMessageID string) error {
	err := r.queries.MarkNotificationSent(ctx, sqlc.MarkNotificationSentParams{
		ID:                id,
		ProviderMessageID: sql.NullString{String: providerMessageID, Valid: providerMessageID != ""},
	})
	if err != nil {
		return fmt.Errorf("mark notification sent: %w", err)
	}

	return nil
}

func (r *NotificationRepository) MarkFailed(ctx context.Context, id uuid.UUID, lastError string) error {
	if len(lastError) > maxErrorLength {
		lastError = lastError[:maxErrorLength]
	}

	err := r.queries.MarkNotificationFailed(ctx, sqlc.MarkNotificationFailedParams{
		ID:        id,
		LastError: sql.NullString{String: lastError, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("mark notification failed: %w", err)
	}

	return nil
}

func toDomain(row sqlc.Notification) *domain.Notification {
	n := &domain.Notification{
		ID:                row.ID,
		EventID:           row.EventID,
		Channel:           row.Channel,
		Template:          row.Template,
		Recipient:         row.Recipient,
		Status:            domain.Status(row.Status),
		Attempts:          int(row.Attempts),
		LastError:         row.LastError.String,
		ProviderMessageID: row.ProviderMessageID.String,
		CreatedAt:         row.CreatedAt,
	}
	if row.SentAt.Valid {
		sentAt := row.SentAt.Time
		n.SentAt = &sentAt
	}

	return n
}
