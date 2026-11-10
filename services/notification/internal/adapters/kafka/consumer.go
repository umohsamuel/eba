package kafka

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
)

type Message struct {
	Topic   string
	Key     []byte
	Value   []byte
	Headers map[string]string
}

type HandlerFunc func(ctx context.Context, msg Message) error

type Consumer struct {
	group sarama.ConsumerGroup
}

func NewKafkaConsumer(brokersUrl []string, groupID string) (*Consumer, error) {
	config := sarama.NewConfig()

	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = true

	g, err := sarama.NewConsumerGroup(brokersUrl, groupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{group: g}, nil
}

func (c *Consumer) Run(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	// With Return.Errors on, sarama blocks if nobody reads this channel.
	go func() {
		for err := range c.group.Errors() {
			slog.Error("kafka consumer group error", "error", err)
		}
	}()

	for {
		if err := c.group.Consume(ctx, topics, handler); err != nil {
			if errors.Is(err, sarama.ErrClosedConsumerGroup) {
				return nil
			}
			slog.Error("kafka consume", "topics", topics, "error", err)

			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
			}
		}

		if ctx.Err() != nil {
			return c.group.Close()
		}
	}
}

type ConsumerGroupHandler struct {
	handle        HandlerFunc
	router        *RetryRouter
	delayDuration time.Duration
}

func NewConsumerGroupHandler(handle HandlerFunc, router *RetryRouter, delayDuration time.Duration) *ConsumerGroupHandler {
	return &ConsumerGroupHandler{
		handle:        handle,
		router:        router,
		delayDuration: delayDuration,
	}
}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()

	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			message := toMessage(msg)

			if failedAt, err := time.Parse(time.RFC3339, message.Headers[headerFailedAt]); err == nil {
				if !WaitUntil(ctx, failedAt.Add(h.delayDuration)) {
					return nil
				}
			}

			if err := h.handle(ctx, message); err != nil {
				if ctx.Err() != nil {
					return nil
				}

				slog.Warn("kafka message failed, routing to next stage", "topic", msg.Topic, "error", err)
				if errRoute := h.router.Route(msg, err); errRoute != nil {
					return errRoute
				}
			}

			session.MarkMessage(msg, "")

		case <-ctx.Done():
			return nil
		}
	}
}

func WaitUntil(ctx context.Context, t time.Time) bool {
	remainingDelay := time.Until(t)
	if remainingDelay <= 0 {
		return true
	}

	slog.Debug("waiting before retry", "delay", remainingDelay)
	timer := time.NewTimer(remainingDelay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}

func toMessage(msg *sarama.ConsumerMessage) Message {
	headers := make(map[string]string, len(msg.Headers))
	for _, header := range msg.Headers {
		headers[string(header.Key)] = string(header.Value)
	}

	return Message{
		Topic:   msg.Topic,
		Key:     msg.Key,
		Value:   msg.Value,
		Headers: headers,
	}
}
