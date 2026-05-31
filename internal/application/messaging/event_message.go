package messaging

/*import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/events"
)

type EventMessage struct {
	ID            string
	Name          string
	AggregateID   string
	AggregateType string
	Version       int
	OccurredAt    time.Time
	Payload       any
}

func ToEventMessage(e events.DomainEvent) EventMessage {
	return EventMessage{
		ID:          e.EventID(),
		Name:        e.EventName(),
		AggregateID: e.AggregateID(),
		//AggregateType: e.AggregateType(),
		Version:    e.Version(),
		OccurredAt: e.OccurredAt(),
		Payload:    e.Payload(),
	}
}
*/