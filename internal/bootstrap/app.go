package bootstrap

import (
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/victorotene80/medilog-api/internal/shared/config"
	"github.com/victorotene80/medilog-api/internal/shared/logging"
)

type App struct {
	Router http.Handler
	DB     any
	Redis  *redis.Client
	Stop   func()
}

func InitializeApp() (*App, error) {
	bootstrapLog, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create bootstrap logger: %w", err)
	}
	defer bootstrapLog.Sync()

	cfg, err := config.Load()
	if err != nil {
		bootstrapLog.Fatal("failed to load config", zap.Error(err))
		return nil, err
	}

	// Telemetry (OTel) — COMMENTED OUT
	// Uncomment to re-enable.
	// tel, err := initializeTelemetry(cfg, bootstrapLog)
	// if err != nil {
	// 	bootstrapLog.Fatal("failed to initialize telemetry", zap.Error(err))
	// 	return nil, err
	// }
	// var otelLP otellog.LoggerProvider
	// if tel != nil {
	// 	otelLP = tel.LoggerProvider
	// }

	logProvider, err := logging.NewLoggerProvider(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logProvider.Sync()

	logger := logProvider.Logger()

	persistenceLayer, err := initializePersistence(cfg, logger)
	if err != nil {
		logger.Error("failed to initialize persistence", zap.Error(err))
		return nil, err
	}

	redisClient, err := initializeRedis(cfg.Redis, logger)
	if err != nil {
		logger.Error("failed to initialize redis", zap.Error(err))
		return nil, err
	}

	httpSvc := initializeHTTPClient(cfg.HTTP)
	externalServices := initializeExternalServices(cfg, httpSvc)

	messagePublisher := initializeMessagePublisher(persistenceLayer, logger)

	// Messaging (Kafka + RabbitMQ) — COMMENTED OUT
	// Uncomment to re-enable.
	// consumers, stopMessaging := initializeMessaging(persistenceLayer, cfg, logger)
	// registerConsumers(consumers, logger)
	// consumeCtx, cancelConsume := context.WithCancel(context.Background())
	// go func() {
	// 	if err := consumers.EventConsumer.Consume(consumeCtx); err != nil {
	// 		logger.Error("event consumer exited", zap.Error(err))
	// 	}
	// }()
	// go func() {
	// 	if err := consumers.TaskConsumer.Consume(consumeCtx); err != nil {
	// 		logger.Error("task consumer exited", zap.Error(err))
	// 	}
	// }()

	commandBus, authSvc := initializeCommands(
		persistenceLayer,
		externalServices,
		messagePublisher,
		cfg,
		redisClient,
		logger,
	)

	router := initializeHTTP(commandBus, logger, authSvc, redisClient)

	stop := func() {
		//cancelConsume()
		//stopMessaging()

		//tel.Shutdown(shutdownCtx)

		logger.Info("app shutdown complete")
		logProvider.Sync()
	}

	return &App{
		Router: router,
		DB:     persistenceLayer.DB,
		Redis:  redisClient,
		Stop:   stop,
	}, nil
}
