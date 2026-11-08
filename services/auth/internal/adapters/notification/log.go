package notification

import (
	"context"
	"log/slog"
)

type LogNotifier struct{}

func (LogNotifier) SendPasswordReset(ctx context.Context, email, token string) error {
	slog.WarnContext(ctx, "password reset requested (development notifier)", "email", email, "token", token)
	return nil
}
