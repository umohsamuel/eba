package kafka

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/IBM/sarama"
)

const (
	headerRetryCount    = "x-retry-count"
	headerFailedAt      = "x-failed-at"
	headerOriginalTopic = "x-original-topic"
	headerError         = "x-error"
)

// permanentError marks a failure that will never succeed on retry, e.g. bad input.
type permanentError struct {
	err error
}

func (e permanentError) Error() string { return e.err.Error() }
func (e permanentError) Unwrap() error { return e.err }

func Permanent(err error) error {
	return permanentError{err: err}
}

func isPermanent(err error) bool {
	var p permanentError
	return errors.As(err, &p)
}

type RetryRouter struct {
	producer    *KafkaProducer
	retryTopics []string
	dlqTopic    string
}

func NewRetryRouter(producer *KafkaProducer, retryTopics []string, dlqTopic string) *RetryRouter {
	return &RetryRouter{
		producer:    producer,
		retryTopics: retryTopics,
		dlqTopic:    dlqTopic,
	}
}

// Schedule sends a new message to the first retry tier, as if it had just failed
// on originalTopic.
func (r *RetryRouter) Schedule(key, value []byte, originalTopic string, cause error) error {
	return r.Route(&sarama.ConsumerMessage{Topic: originalTopic, Key: key, Value: value}, cause)
}

func (r *RetryRouter) Route(msg *sarama.ConsumerMessage, cause error) error {
	retryCount := 0
	originalTopic := msg.Topic

	headers := make([]sarama.RecordHeader, 0, len(msg.Headers)+4)
	for _, header := range msg.Headers {
		switch string(header.Key) {
		case headerRetryCount:
			retryCount, _ = strconv.Atoi(string(header.Value))
		case headerOriginalTopic:
			originalTopic = string(header.Value)
		case headerFailedAt, headerError:
		default:
			headers = append(headers, *header)
		}
	}

	nextTopic := r.dlqTopic
	if !isPermanent(cause) && retryCount >= 0 && retryCount < len(r.retryTopics) {
		nextTopic = r.retryTopics[retryCount]
	}

	headers = append(headers,
		sarama.RecordHeader{Key: []byte(headerRetryCount), Value: []byte(strconv.Itoa(retryCount + 1))},
		sarama.RecordHeader{Key: []byte(headerFailedAt), Value: []byte(time.Now().Format(time.RFC3339))},
		sarama.RecordHeader{Key: []byte(headerOriginalTopic), Value: []byte(originalTopic)},
		sarama.RecordHeader{Key: []byte(headerError), Value: []byte(cause.Error())},
	)

	slog.Warn("routing kafka message", "to", nextTopic, "attempt", retryCount+1, "error", cause)
	return r.producer.send(&sarama.ProducerMessage{
		Topic:   nextTopic,
		Key:     sarama.ByteEncoder(msg.Key),
		Value:   sarama.ByteEncoder(msg.Value),
		Headers: headers,
	})
}
