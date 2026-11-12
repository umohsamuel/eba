package outbox

import "github.com/google/uuid"

type OutboxEvent struct {
	ID            uuid.UUID
	AggregateType string
	AggregateID   string
	Type          string
	Payload       []byte
}
