package kafka

import "github.com/umohsamuel/eba/pkg/events"

const (
	TopicRequested = events.TopicNotificationRequested
	TopicRetry1    = events.TopicNotificationRequested + ".retry.1"
	TopicRetry2    = events.TopicNotificationRequested + ".retry.2"
	TopicRetry3    = events.TopicNotificationRequested + ".retry.3"
	TopicDLQ       = events.TopicNotificationRequested + ".dlq"
)

var RetryTopics = []string{TopicRetry1, TopicRetry2, TopicRetry3}

var AllTopics = []string{TopicRequested, TopicRetry1, TopicRetry2, TopicRetry3, TopicDLQ}

const (
	GroupMain   = "notification.main"
	GroupRetry1 = "notification.retry.1"
	GroupRetry2 = "notification.retry.2"
	GroupRetry3 = "notification.retry.3"
)

var RetryGroups = []string{GroupRetry1, GroupRetry2, GroupRetry3}
