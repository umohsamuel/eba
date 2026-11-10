package kafka

import (
	"log/slog"

	"github.com/IBM/sarama"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
}

func NewKafkaProducer(brokersUrl []string) (*KafkaProducer, error) {
	config := sarama.NewConfig()

	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Idempotent = true
	config.Producer.Retry.Max = 3
	config.Net.MaxOpenRequests = 1

	p, err := sarama.NewSyncProducer(brokersUrl, config)
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{producer: p}, nil
}

func (kp *KafkaProducer) SendMessage(topic, key, value string) error {
	return kp.send(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	})
}

func (kp *KafkaProducer) send(msg *sarama.ProducerMessage) error {
	partition, offset, err := kp.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	slog.Debug("kafka message stored", "topic", msg.Topic, "partition", partition, "offset", offset)
	return nil
}

// Close flushes anything in flight and releases the connection to the brokers.
func (kp *KafkaProducer) Close() error {
	return kp.producer.Close()
}
