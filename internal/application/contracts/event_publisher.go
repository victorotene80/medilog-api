package contracts

import (
	"context"

	appmsg "github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/domain/events"
)

type MessagePublisher interface {
	Publish(ctx context.Context, domainEvents []events.DomainEvent, meta map[string]string) error
	PublishEnvelope(ctx context.Context, envelope appmsg.Envelope) error
}