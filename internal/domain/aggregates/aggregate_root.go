package aggregates

import "github.com/victorotene80/medilog-api/internal/domain/events"

type AggregateRoot struct {
	id                int64
	version           int
	uncommittedEvents []events.DomainEvent
}

func NewAggregateRoot(id int64, version int) *AggregateRoot {
	return &AggregateRoot{
		id:                id,
		version:           version,
		uncommittedEvents: make([]events.DomainEvent, 0),
	}
}

func (a *AggregateRoot) ID() int64         { return a.id }
func (a *AggregateRoot) SetID(id int64)    { a.id = id }
func (a *AggregateRoot) Version() int      { return a.version }
func (a *AggregateRoot) SetVersion(v int)  { a.version = v }

func (a *AggregateRoot) RaiseEvent(event events.DomainEvent) {
	a.uncommittedEvents = append(a.uncommittedEvents, event)
}

func (a *AggregateRoot) PullEvents() []events.DomainEvent {
	cp := make([]events.DomainEvent, len(a.uncommittedEvents))
	copy(cp, a.uncommittedEvents)
	return cp
}

func (a *AggregateRoot) ClearEvents() {
	a.uncommittedEvents = a.uncommittedEvents[:0]
}

func (a *AggregateRoot) CommitVersion(newVersion int) {
	a.version = newVersion
}