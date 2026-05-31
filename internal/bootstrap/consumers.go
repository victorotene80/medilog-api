package bootstrap

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"go.uber.org/zap"
)

func registerConsumers(c Consumers, logger *zap.Logger) {
	// Kafka integration events
	c.EventConsumer.Subscribe("auth.payment.completed.v1", func(ctx context.Context, env messaging.Envelope) error {
		// handle integration event from another service
		return nil
	})

	// RabbitMQ internal tasks
	c.TaskConsumer.Subscribe("auth.send-welcome-email.v1", func(ctx context.Context, env messaging.Envelope) error {
		// handle background task
		return nil
	})

	logger.Info("consumers registered")
}

/*package bootstrap

import (
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
	// ── EventConsumer: Kafka ──────────────────────────────────────────────────
	// Subscribe to domain events produced by OTHER services.
	// Uncomment and add handlers as you integrate with other services.

	// c.EventConsumer.Subscribe(
	// 	"payments.PaymentCompleted",
	// 	func(ctx context.Context, env messaging.Envelope) error {
	// 		return handlers.HandlePaymentCompleted(ctx, env, logger)
	// 	},
	// )

	// ── TaskConsumer: RabbitMQ ────────────────────────────────────────────────
	// Subscribe to internal retry queues and dead-letter replay.

	// c.TaskConsumer.Subscribe(
	// 	"auth.send-welcome-email",
	// 	func(ctx context.Context, env messaging.Envelope) error {
	// 		return handlers.HandleSendWelcomeEmail(ctx, env, logger)
	// 	},
	// )

	logger.Info("consumers registered")
}*/
