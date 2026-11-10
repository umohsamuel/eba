package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/umohsamuel/eba/pkg/events"
	domain "github.com/umohsamuel/eba/services/notification/internal/domain/notification"
)

type NotificationHandler interface {
	Handle(ctx context.Context, req events.NotificationRequested) error
}

func HandleNotificationRequested(h NotificationHandler) HandlerFunc {
	return func(ctx context.Context, msg Message) error {
		var req events.NotificationRequested
		if err := json.Unmarshal(msg.Value, &req); err != nil {
			return Permanent(fmt.Errorf("decode %s: %w", msg.Topic, err))
		}

		if err := h.Handle(ctx, req); err != nil {
			if domain.IsPermanent(err) {
				return Permanent(err)
			}
			return err
		}

		return nil
	}
}
