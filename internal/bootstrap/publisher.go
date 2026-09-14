package bootstrap

import (
	"context"

	"go.uber.org/zap"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	appmsg "github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/domain/events"
	outboxContracts "github.com/victorotene80/medilog-api/internal/infrastructure/messaging/outbox"
	outboxPublisher "github.com/victorotene80/medilog-api/internal/infrastructure/messaging/outbox"
)

// initializeMessagePublisher returns the outbox-backed publisher, or a no-op
// when messaging is disabled.
//
// MESSAGING_ENABLED defaults to false, and the publisher used to be wired
// regardless: every registration and login wrote an outbox_events row that no
// relay was running to drain. The table grew monotonically, and switching
// messaging on later would replay the entire accumulated history in batches —
// months of user.created events delivered as though they had just happened.
func newEventPublisher(
	outboxRepo outboxContracts.OutboxRepository,
	enabled bool,
	logger *zap.Logger,
) appContracts.MessagePublisher {
	if !enabled {
		logger.Info("messaging disabled — domain events are not recorded to the outbox")
		return noopPublisher{}
	}

	if outboxRepo == nil {
		logger.Fatal("outbox repository is nil — cannot initialize message publisher")
	}

	logger.Info("message publisher ready (outbox-backed)")

	return outboxPublisher.NewPublisher(outboxRepo)
}

var _ appContracts.MessagePublisher = (*noopPublisher)(nil)

type noopPublisher struct{}

func (noopPublisher) Publish(context.Context, []events.DomainEvent, map[string]string) error {
	return nil
}

func (noopPublisher) PublishEnvelope(context.Context, appmsg.Envelope) error {
	return nil
}
