package email

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "github.com/umohsamuel/eba/services/notification/internal/domain/notification"
	service "github.com/umohsamuel/eba/services/notification/internal/services/notification"
	"github.com/wneessen/go-mail"
)

type SMTPConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	From      string
	TLSPolicy string
}

type SMTPSender struct {
	client *mail.Client
	from   string
}

func NewSMTPSender(cfg SMTPConfig) (*SMTPSender, error) {
	policy, err := tlsPolicy(cfg.TLSPolicy)
	if err != nil {
		return nil, err
	}

	opts := []mail.Option{
		mail.WithPort(cfg.Port),
		mail.WithTLSPolicy(policy),
		mail.WithTimeout(15 * time.Second),
	}
	if cfg.Username != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
			mail.WithUsername(cfg.Username),
			mail.WithPassword(cfg.Password),
		)
	}

	client, err := mail.NewClient(cfg.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("create smtp client: %w", err)
	}

	return &SMTPSender{client: client, from: cfg.From}, nil
}

func (s *SMTPSender) Send(ctx context.Context, e service.Email) (string, error) {
	msg := mail.NewMsg()
	if err := msg.From(s.from); err != nil {
		return "", fmt.Errorf("set from: %w", err)
	}
	if err := msg.To(e.To); err != nil {
		return "", fmt.Errorf("%w: %v", domain.ErrInvalidRequest, err)
	}

	msg.Subject(e.Subject)
	msg.SetMessageID()
	msg.SetBodyString(mail.TypeTextPlain, e.Text)
	msg.AddAlternativeString(mail.TypeTextHTML, e.HTML)

	if err := s.client.DialAndSendWithContext(ctx, msg); err != nil {
		var sendErr *mail.SendError
		if errors.As(err, &sendErr) && sendErr.Reason == mail.ErrSMTPRcptTo && !sendErr.IsTemp() {
			return "", fmt.Errorf("%w: %v", domain.ErrRecipientRejected, err)
		}
		return "", err
	}

	return msg.GetMessageID(), nil
}

func tlsPolicy(name string) (mail.TLSPolicy, error) {
	switch name {
	case "", "mandatory":
		return mail.TLSMandatory, nil
	case "opportunistic":
		return mail.TLSOpportunistic, nil
	case "none":
		return mail.NoTLS, nil
	default:
		return 0, fmt.Errorf("unknown SMTP TLS policy %q", name)
	}
}
