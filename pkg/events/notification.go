package events

import "time"

const TopicNotificationRequested = "notifications.requested"

const ChannelEmail = "email"

const (
	TemplateAuthWelcome         = "auth.welcome"
	TemplateAuthVerifyEmail     = "auth.verify_email"
	TemplateAuthPasswordReset   = "auth.password_reset"
	TemplateAuthPasswordChanged = "auth.password_changed"
)

type NotificationRequested struct {
	EventID    string            `json:"event_id"`
	Channel    string            `json:"channel"`
	Template   string            `json:"template"`
	Recipient  string            `json:"recipient"`
	Data       map[string]string `json:"data,omitempty"`
	OccurredAt time.Time         `json:"occurred_at"`
}
