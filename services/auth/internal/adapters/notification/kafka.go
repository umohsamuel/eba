package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/umohsamuel/eba/pkg/events"
)

type KafkaNotifier struct {
	producer         sarama.SyncProducer
	passwordResetURL string
	passwordResetTTL time.Duration
}

func NewKafkaNotifier(brokers []string, passwordResetURL string, passwordResetTTL time.Duration) (*KafkaNotifier, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Idempotent = true
	config.Producer.Retry.Max = 3
	config.Net.MaxOpenRequests = 1

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}

	return &KafkaNotifier{
		producer:         producer,
		passwordResetURL: passwordResetURL,
		passwordResetTTL: passwordResetTTL,
	}, nil
}

func (n *KafkaNotifier) SendPasswordReset(ctx context.Context, email, token string) error {
	return n.publish(events.NotificationRequested{
		EventID:   uuid.NewString(),
		Channel:   events.ChannelEmail,
		Template:  events.TemplateAuthPasswordReset,
		Recipient: email,
		Data: map[string]string{
			"reset_url":  n.passwordResetURL + "#token=" + token,
			"expires_in": humanize(n.passwordResetTTL),
		},
		OccurredAt: time.Now().UTC(),
	})
}

func (n *KafkaNotifier) Close() error {
	return n.producer.Close()
}

func (n *KafkaNotifier) publish(evt events.NotificationRequested) error {
	value, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("encode notification event: %w", err)
	}

	_, _, err = n.producer.SendMessage(&sarama.ProducerMessage{
		Topic: events.TopicNotificationRequested,
		Key:   sarama.StringEncoder(evt.Recipient),
		Value: sarama.ByteEncoder(value),
	})
	if err != nil {
		return fmt.Errorf("publish notification event: %w", err)
	}

	return nil
}

func humanize(d time.Duration) string {
	switch {
	case d >= time.Hour && d%time.Hour == 0:
		return plural(int(d/time.Hour), "hour")
	case d >= time.Minute:
		return plural(int(d/time.Minute), "minute")
	default:
		return plural(int(d/time.Second), "second")
	}
}

func plural(n int, unit string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", unit)
	}
	return fmt.Sprintf("%d %ss", n, unit)
}
