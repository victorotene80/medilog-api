package bootstrap

import (
	appmsg "github.com/victorotene80/medilog-api/internal/application/messaging"
	"go.uber.org/zap"
)

// registerConsumers is the single place that maps topic/queue names to
// application handlers. It mirrors commands.go, which does the same for
// the command bus.
//
// Rules:
//   - One Subscribe call per topic/queue.
//   - No business logic here — only wiring.
func registerConsumers(c Consumers, logger *zap.Logger) {
	// ── EventConsumer: RabbitMQ ─────────────────────────────────────────────
	// Subscribe to domain events produced by OTHER services.
	c.EventConsumer.Subscribe(appmsg.EventUserCreated, HandleAuthUserCreated(logger))
	c.EventConsumer.Subscribe(appmsg.EventUserLocked, HandleAuthUserLocked(logger))
	c.EventConsumer.Subscribe(appmsg.EventSessionCreated, HandleAuthSessionCreated(logger))
	c.EventConsumer.Subscribe(appmsg.EventSessionRevoked, HandleAuthSessionRevoked(logger))
	c.EventConsumer.Subscribe(appmsg.EventPasswordChanged, HandleAuthPasswordChanged(logger))

	// ── TaskConsumer: RabbitMQ ────────────────────────────────────────────────
	// Subscribe to internal retry queues and dead-letter replay.
	c.TaskConsumer.Subscribe(appmsg.TaskSendWelcomeEmail, HandleSendWelcomeEmail(logger))
	c.TaskConsumer.Subscribe(appmsg.TaskSendVerification, HandleSendVerification(logger))
	c.TaskConsumer.Subscribe(appmsg.TaskSyncAnalyticsUser, HandleSyncAnalyticsUser(logger))

	logger.Info("consumers registered",
		zap.Int("event_handlers", 5),
		zap.Int("task_handlers", 3),
	)
}
