package outbox

import (
	"context"
	"fmt"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	appmsg "github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/domain/events"
)

var _ appContracts.MessagePublisher = (*Publisher)(nil)

type Publisher struct {
	repo OutboxRepository
}

func NewPublisher(repo OutboxRepository) *Publisher {
	return &Publisher{repo: repo}
}

func (p *Publisher) Publish(
	ctx context.Context,
	domainEvents []events.DomainEvent,
	meta map[string]string,
) error {
	for _, evt := range domainEvents {
		env, err := appmsg.ToEnvelope(evt, meta)
		if err != nil {
			return fmt.Errorf("outbox publisher: to envelope: %w", err)
		}

		if err := p.repo.Add(ctx, env); err != nil {
			return fmt.Errorf("outbox publisher: insert envelope: %w", err)
		}
	}
	return nil
}

func (p *Publisher) PublishEnvelope(ctx context.Context, envelope appmsg.Envelope) error {
	return p.repo.Add(ctx, envelope)
}