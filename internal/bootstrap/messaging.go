package bootstrap

import (
	"context"

	coremsg "github.com/victorotene80/medilog-api/internal/infrastructure/messaging"
	kafkaInfra "github.com/victorotene80/medilog-api/internal/infrastructure/messaging/kafka"
	rabbitInfra "github.com/victorotene80/medilog-api/internal/infrastructure/messaging/rabbitmq"
	"github.com/victorotene80/medilog-api/internal/shared/config"
	"go.uber.org/zap"
)

type Consumers struct {
	EventConsumer coremsg.MessageConsumer
	TaskConsumer  coremsg.MessageConsumer
}

func initializeMessaging(
	p *Persistence,
	cfg *config.Config,
	logger *zap.Logger,
) (Consumers, func()) {
	//eventBroker := kafkaInfra.NewBroker(kafkaInfra.BrokerConfig{
	//	Brokers:      cfg.Messaging.Kafka.Brokers,
	//	TopicPrefix:  cfg.Messaging.Kafka.TopicPrefix,
	//	WriteTimeout: cfg.Messaging.Kafka.WriteTimeout,
	//})
	//logger.Info("event broker ready (kafka)", zap.Strings("brokers", cfg.Messaging.Kafka.Brokers))

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

	logger.Info("task broker ready (rabbitmq)",
		zap.String("exchange", cfg.Messaging.RabbitMQ.Exchange),
		zap.String("retry_exchange", cfg.Messaging.RabbitMQ.RetryExchange),
		zap.String("dl_exchange", cfg.Messaging.RabbitMQ.DLExchange),
	)

	relay := coremsg.New(
		p.OutboxRepo,
		coremsg.Brokers{
			//EventBroker: eventBroker,
			TaskBroker: taskBroker,
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

	eventConsumer := kafkaInfra.NewConsumer(kafkaInfra.ConsumerConfig{
		Brokers:     cfg.Messaging.Kafka.Brokers,
		GroupID:     cfg.Messaging.Kafka.ConsumerGroupID,
		TopicPrefix: cfg.Messaging.Kafka.TopicPrefix,
	})

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

		if err := eventConsumer.Close(); err != nil {
			logger.Error("error closing event consumer", zap.Error(err))
		}
		if err := taskConsumer.Close(); err != nil {
			logger.Error("error closing task consumer", zap.Error(err))
		}
		//if err := eventBroker.Close(); err != nil {
		//	logger.Error("error closing event broker", zap.Error(err))
		//}
		if err := taskBroker.Close(); err != nil {
			logger.Error("error closing task broker", zap.Error(err))
		}

		logger.Info("messaging shutdown complete")
	}

	return Consumers{
		EventConsumer: eventConsumer,
		TaskConsumer:  taskConsumer,
	}, stop
}
