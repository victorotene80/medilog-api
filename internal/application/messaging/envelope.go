package messaging

import (
	"encoding/json"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/events"
)

type Kind string

const (
	KindIntegrationEvent Kind = "integration_event"
	KindTask             Kind = "task"
)

type Envelope struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Kind          Kind              `json:"kind"`
	AggregateID   int64             `json:"aggregate_id"`
	AggregateType string            `json:"aggregate_type"`
	OccurredAt    time.Time         `json:"occurred_at"`
	Payload       []byte            `json:"payload"`
	Metadata      map[string]string `json:"metadata"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	CausationID   string            `json:"causation_id,omitempty"`
	Version       int               `json:"version"`
}

func ToEnvelope(e events.DomainEvent, metadata map[string]string) (Envelope, error) {
	b, err := json.Marshal(e.Payload())
	if err != nil {
		return Envelope{}, err
	}

	kind := Kind(metadata["message_kind"])
	if kind == "" {
		kind = KindIntegrationEvent
	}

	// Explicit metadata wins; otherwise resolve the domain event name to its
	// published routing key. Falling straight back to e.EventName() is what made
	// register's user.created land on "user.created", which no queue binds to —
	// the exchange discarded it and the relay marked the row sent.
	name := metadata["message_name"]
	if name == "" {
		if wire, ok := WireNameFor(e.EventName()); ok {
			name = wire
		} else {
			name = e.EventName()
		}
	}

	aggregateType := metadata["aggregate_type"]

	return Envelope{
		ID:            e.EventID(),
		Name:          name,
		Kind:          kind,
		AggregateID:   e.AggregateID(),
		AggregateType: aggregateType,
		OccurredAt:    e.OccurredAt(),
		Payload:       b,
		Metadata:      metadata,
		CorrelationID: metadata["correlation_id"],
		CausationID:   metadata["causation_id"],
		Version:       e.Version(),
	}, nil
}
