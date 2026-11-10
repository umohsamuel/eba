package notification

import (
	"context"
	"fmt"
	"log/slog"
	"net/mail"

	"github.com/google/uuid"
	"github.com/umohsamuel/eba/pkg/events"
	domain "github.com/umohsamuel/eba/services/notification/internal/domain/notification"
)

type Email struct {
	To      string
	Subject string
	HTML    string
	Text    string
}

type EmailSender interface {
	Send(ctx context.Context, email Email) (messageID string, err error)
}

type Rendered struct {
	Subject string
	HTML    string
	Text    string
}

type Renderer interface {
	Render(template string, data map[string]string) (*Rendered, error)
}

type Service struct {
	repo     domain.Repository
	renderer Renderer
	email    EmailSender
}

type Config struct {
	Repository domain.Repository
	Renderer   Renderer
	Email      EmailSender
}

func New(cfg Config) *Service {
	return &Service{
		repo:     cfg.Repository,
		renderer: cfg.Renderer,
		email:    cfg.Email,
	}
}

func (s *Service) Handle(ctx context.Context, req events.NotificationRequested) error {
	if err := validate(req); err != nil {
		return err
	}

	rendered, err := s.renderer.Render(req.Template, req.Data)
	if err != nil {
		return err
	}

	n, err := s.repo.CreateIfAbsent(ctx, &domain.Notification{
		ID:        uuid.New(),
		EventID:   req.EventID,
		Channel:   req.Channel,
		Template:  req.Template,
		Recipient: req.Recipient,
		Status:    domain.StatusPending,
	})
	if err != nil {
		return fmt.Errorf("record notification: %w", err)
	}

	if n.Status == domain.StatusSent {
		slog.InfoContext(ctx, "notification already sent, skipping", "event_id", n.EventID)
		return nil
	}

	messageID, err := s.email.Send(ctx, Email{
		To:      req.Recipient,
		Subject: rendered.Subject,
		HTML:    rendered.HTML,
		Text:    rendered.Text,
	})
	if err != nil {
		if markErr := s.repo.MarkFailed(ctx, n.ID, err.Error()); markErr != nil {
			slog.ErrorContext(ctx, "mark notification failed", "event_id", n.EventID, "error", markErr)
		}
		return fmt.Errorf("send email: %w", err)
	}

	if err := s.repo.MarkSent(ctx, n.ID, messageID); err != nil {
		slog.ErrorContext(ctx, "mark notification sent", "event_id", n.EventID, "error", err)
	}

	slog.InfoContext(ctx, "notification sent", "event_id", n.EventID, "template", n.Template, "channel", n.Channel)
	return nil
}

func validate(req events.NotificationRequested) error {
	if req.EventID == "" {
		return fmt.Errorf("%w: missing event_id", domain.ErrInvalidRequest)
	}

	if req.Channel != events.ChannelEmail {
		return fmt.Errorf("%w: %q", domain.ErrUnknownChannel, req.Channel)
	}

	if _, err := mail.ParseAddress(req.Recipient); err != nil {
		return fmt.Errorf("%w: invalid recipient", domain.ErrInvalidRequest)
	}

	return nil
}
