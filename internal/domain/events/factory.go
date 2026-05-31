package events

import (
	"time"

	"github.com/google/uuid"
)

func NewEvent(
	name string,
	aggregateID int64,
	payload any,
	meta map[string]string,
) DomainEvent {
	if meta == nil {
		meta = map[string]string{}
	}
	return baseDomainEvent{
		id:          uuid.NewString(),
		name:        name,
		aggregateID: aggregateID,
		payload:     payload,
		meta:        meta,
		version:     1,
		occurredAt:  time.Now().UTC(),
	}
}
