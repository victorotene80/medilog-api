package bootstrap

import (
	"context"
	"time"

	coremsg "github.com/victorotene80/medilog-api/internal/infrastructure/messaging"
	rabbitInfra "github.com/victorotene80/medilog-api/internal/infrastructure/messaging/rabbitmq"
	"github.com/victorotene80/medilog-api/internal/shared/config"
	"go.uber.org/zap"
)

type Consumers struct {
	EventConsumer coremsg.MessageConsumer
	TaskConsumer  coremsg.MessageConsumer
}

// relayShutdownGrace bounds how long shutdown waits for an in-flight relay tick
// to finish publishing before the brokers are closed underneath it.
const relayShutdownGrace = 10 * time.Second

func initializeMessaging(
	p *Persistence,
	cfg *config.Config,
	logger *zap.Logger,
) (Consumers, func()) {
	if !cfg.Messaging.Enabled {
		logger.Info("messaging disabled — skipping Kafka + RabbitMQ init")
		return Consumers{}, func() {}
	}

	eventBroker, err := rabbitInfra.NewBroker(rabbitInfra.BrokerConfig{
		DSN:            cfg.Messaging.RabbitMQ.DSN,
		Exchange:       cfg.Messaging.RabbitMQ.Exchange,
		RetryExchange:  cfg.Messaging.RabbitMQ.RetryExchange,
		DLExchange:     cfg.Messaging.RabbitMQ.DLExchange,
		PublishTimeout: cfg.Messaging.RabbitMQ.PublishTimeout,
	})
	if err != nil {
		logger.Fatal("failed to create event broker (rabbitmq)", zap.Error(err))
	}
	logger.Info("event broker ready (rabbitmq)", zap.String("exchange", cfg.Messaging.RabbitMQ.Exchange))

	taskBroker, err := rabbitInfra.NewBroker(rabbitInfra.BrokerConfig{
		DSN:            cfg.Messaging.RabbitMQ.DSN,
		Exchange:       cfg.Messaging.RabbitMQ.Exchange,
		RetryExchange:  cfg.Messaging.RabbitMQ.RetryExchange,
		DLExchange:     cfg.Messaging.RabbitMQ.DLExchange,
		PublishTimeout: cfg.Messaging.RabbitMQ.PublishTimeout,
	})
	if err != nil {
		logger.Fatal("failed to create task broker (rabbitmq)", zap.Error(err))
	}
	logger.Info("task broker ready (rabbitmq)", zap.String("exchange", cfg.Messaging.RabbitMQ.Exchange))

	relay := coremsg.New(
		p.OutboxRepo,
		coremsg.Brokers{
			EventBroker: eventBroker,
			TaskBroker:  taskBroker,
		},
		coremsg.Config{
			PollInterval:       cfg.Messaging.Relay.PollInterval,
			BatchSize:          cfg.Messaging.Relay.BatchSize,
			ReclaimAfter:       cfg.Messaging.Relay.ReclaimAfter,
			DefaultEventBroker: cfg.Messaging.Relay.DefaultEventBroker,
			DefaultTaskBroker:  cfg.Messaging.Relay.DefaultTaskBroker,
			EventRoutes:        cfg.Messaging.Relay.EventRoutes,
			TaskRoutes:         cfg.Messaging.Relay.TaskRoutes,
		},
		logger,
	)
	relayCtx, cancelRelay := context.WithCancel(context.Background())
	go relay.Run(relayCtx)

	eventConsumer, err := rabbitInfra.NewConsumer(rabbitInfra.ConsumerConfig{
		DSN:           cfg.Messaging.RabbitMQ.DSN,
		Exchange:      cfg.Messaging.RabbitMQ.Exchange,
		RetryExchange: cfg.Messaging.RabbitMQ.RetryExchange,
		DLExchange:    cfg.Messaging.RabbitMQ.DLExchange,
		MaxRetries:    cfg.Messaging.RabbitMQ.MaxRetries,
		RetryDelay:    cfg.Messaging.RabbitMQ.RetryDelay,
		Prefetch:      10,
	})
	if err != nil {
		logger.Fatal("failed to create event consumer (rabbitmq)", zap.Error(err))
	}

	taskConsumer, err := rabbitInfra.NewConsumer(rabbitInfra.ConsumerConfig{
		DSN:           cfg.Messaging.RabbitMQ.DSN,
		Exchange:      cfg.Messaging.RabbitMQ.Exchange,
		RetryExchange: cfg.Messaging.RabbitMQ.RetryExchange,
		DLExchange:    cfg.Messaging.RabbitMQ.DLExchange,
		MaxRetries:    cfg.Messaging.RabbitMQ.MaxRetries,
		RetryDelay:    cfg.Messaging.RabbitMQ.RetryDelay,
		Prefetch:      10,
	})
	if err != nil {
		logger.Fatal("failed to create task consumer (rabbitmq)", zap.Error(err))
	}

	stop := func() {
		cancelRelay()

		// Wait for the relay to leave its tick before closing the brokers under
		// it. Cancelling and closing in the same breath could tear the AMQP
		// channel down mid-publish for a row already marked in_progress, and the
		// MarkFailed that followed ran on the same cancelled context — leaving
		// the row stuck until the reclaim sweep two minutes into the next boot.
		select {
		case <-relay.Done():
		case <-time.After(relayShutdownGrace):
			logger.Warn("outbox relay did not stop within the shutdown grace period",
				zap.Duration("grace", relayShutdownGrace),
			)
		}

		if err := eventBroker.Close(); err != nil {
			logger.Error("error closing event broker", zap.Error(err))
		}
		if err := taskBroker.Close(); err != nil {
			logger.Error("error closing task broker", zap.Error(err))
		}
		if err := eventConsumer.Close(); err != nil {
			logger.Error("error closing event consumer", zap.Error(err))
		}
		if err := taskConsumer.Close(); err != nil {
			logger.Error("error closing task consumer", zap.Error(err))
		}

		logger.Info("messaging shutdown complete")
	}

	return Consumers{
		EventConsumer: eventConsumer,
		TaskConsumer:  taskConsumer,
	}, stop
}
