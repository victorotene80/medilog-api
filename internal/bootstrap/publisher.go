package bootstrap

import (
	"go.uber.org/zap"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	outboxPublisher "github.com/victorotene80/medilog-api/internal/infrastructure/messaging/outbox"
)

func initializeMessagePublisher(
	p *Persistence,
	logger *zap.Logger,
) appContracts.MessagePublisher {
	if p.OutboxRepo == nil {
		logger.Fatal("outbox repository is nil — cannot initialize message publisher")
	}
	logger.Info("message publisher ready (outbox-backed)")
	return outboxPublisher.NewPublisher(p.OutboxRepo)
}

/*import (
	"context"

	"go.uber.org/zap"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	kafkaInfra "github.com/victorotene80/medilog-api/internal/infrastructure/messaging/kafka"
	outboxPublisher "github.com/victorotene80/medilog-api/internal/infrastructure/messaging/outbox"
	rabbitInfra "github.com/victorotene80/medilog-api/internal/infrastructure/messaging/rabbitmq"
	"github.com/victorotene80/medilog-api/internal/infrastructure/messaging"
	"github.com/victorotene80/medilog-api/internal/shared/config"
)

type Consumers struct {
	EventConsumer messaging.MessageConsumer
	TaskConsumer  messaging.MessageConsumer
}

func initializeMessagePublisher(
	p *Persistence,
	logger *zap.Logger,
) appContracts.MessagePublisher {
	if p.OutboxRepo == nil {
		logger.Fatal("outbox repository is nil — cannot initialize message publisher")
	}
	logger.Info("message publisher ready (outbox-backed)")
	return outboxPublisher.NewPublisher(p.OutboxRepo)
}

func initializeMessaging(
	p *Persistence,
	cfg *config.Config,
	logger *zap.Logger,
) (Consumers, func()) {

	eventBroker := kafkaInfra.NewBroker(kafkaInfra.BrokerConfig{
		Brokers:      cfg.Messaging.Kafka.Brokers,
		TopicPrefix:  cfg.Messaging.Kafka.TopicPrefix,
		WriteTimeout: cfg.Messaging.Kafka.WriteTimeout,
	})
	logger.Info("event broker ready (kafka)", zap.Strings("brokers", cfg.Messaging.Kafka.Brokers))

	taskBroker, err := rabbitInfra.NewBroker(rabbitInfra.BrokerConfig{
		DSN:        cfg.Messaging.RabbitMQ.DSN,
		Exchange:   cfg.Messaging.RabbitMQ.Exchange,
		DLExchange: cfg.Messaging.RabbitMQ.DLExchange,
	})
	if err != nil {
		logger.Fatal("failed to create task broker (rabbitmq)", zap.Error(err))
	}
	logger.Info("task broker ready (rabbitmq)", zap.String("exchange", cfg.Messaging.RabbitMQ.Exchange))

	r := messaging.New(
		p.OutboxRepo,
		messaging.Brokers{
			EventBroker: eventBroker,
			TaskBroker:  taskBroker,
		},
		messaging.Config{
			PollInterval: cfg.Messaging.Relay.PollInterval,
			BatchSize:    cfg.Messaging.Relay.BatchSize,
			BrokerRoutes: cfg.Messaging.Relay.BrokerRoutes,
		},
		logger,
	)
	relayCtx, cancelRelay := context.WithCancel(context.Background())
	go r.Run(relayCtx)

	eventConsumer := kafkaInfra.NewConsumer(kafkaInfra.ConsumerConfig{
		Brokers:     cfg.Messaging.Kafka.Brokers,
		GroupID:     cfg.Messaging.Kafka.ConsumerGroupID,
		TopicPrefix: cfg.Messaging.Kafka.TopicPrefix,
	})

	taskConsumer, err := rabbitInfra.NewConsumer(rabbitInfra.ConsumerConfig{
		DSN:        cfg.Messaging.RabbitMQ.DSN,
		Exchange:   cfg.Messaging.RabbitMQ.Exchange,
		DLExchange: cfg.Messaging.RabbitMQ.DLExchange,
		MaxRetries: cfg.Messaging.RabbitMQ.MaxRetries,
		RetryDelay: cfg.Messaging.RabbitMQ.RetryDelay,
	})
	if err != nil {
		logger.Fatal("failed to create task consumer (rabbitmq)", zap.Error(err))
	}

	consumers := Consumers{
		EventConsumer: eventConsumer,
		TaskConsumer:  taskConsumer,
	}

	stop := func() {
		cancelRelay()
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

	return consumers, stop
}*/