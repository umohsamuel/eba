package kafka

import (
	"errors"
	"fmt"

	"github.com/IBM/sarama"
)

func EnsureTopics(brokers []string, topics []string, partitions int32, replicationFactor int16) error {
	admin, err := sarama.NewClusterAdmin(brokers, sarama.NewConfig())
	if err != nil {
		return fmt.Errorf("connect kafka admin: %w", err)
	}
	defer admin.Close()

	for _, topic := range topics {
		err := admin.CreateTopic(topic, &sarama.TopicDetail{
			NumPartitions:     partitions,
			ReplicationFactor: replicationFactor,
		}, false)
		if err != nil && !errors.Is(err, sarama.ErrTopicAlreadyExists) {
			return fmt.Errorf("create topic %s: %w", topic, err)
		}
	}

	return nil
}
