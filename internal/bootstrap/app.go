package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	otellog "go.opentelemetry.io/otel/log"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/victorotene80/medilog-api/internal/shared/config"
	"github.com/victorotene80/medilog-api/internal/shared/logging"
)

type App struct {
	Router http.Handler
	DB     *gorm.DB
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

	var tel *Telemetry
	if cfg.Telemetry.Enabled {
		tel, err = initializeTelemetry(cfg, bootstrapLog)
		if err != nil {
			bootstrapLog.Fatal("failed to initialize telemetry", zap.Error(err))
			return nil, err
		}
	} else {
		bootstrapLog.Info("telemetry disabled — skipping OTEL SDK init")
	}

	var otelLP otellog.LoggerProvider
	if tel != nil {
		otelLP = tel.LoggerProvider
	}

	logProvider, err := logging.NewLoggerProvider(otelLP)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logProvider.Sync()

	logger := logProvider.Logger()

	// zap's package-level logger is a no-op until it is replaced. The REST
	// handlers log through zap.L() (logAndRespond, request validation, the health
	// probe), so without this the "log the raw error, return a curated message"
	// contract silently discarded the logging half.
	zap.ReplaceGlobals(logger)

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

	// Built during persistence init, because the aggregate repositories need it
	// at construction time to drain domain events onto their own transaction.
	messagePublisher := persistenceLayer.EventPublisher

	var (
		consumers     Consumers
		stopMessaging func()
	)
	if cfg.Messaging.Enabled {
		consumers, stopMessaging = initializeMessaging(persistenceLayer, cfg, logger)
		registerConsumers(consumers, logger)

		consumerCtx, consumerCancel := context.WithCancel(context.Background())
		go func() {
			if err := consumers.EventConsumer.Consume(consumerCtx); err != nil {
				logger.Error("event consumer exited", zap.Error(err))
			}
		}()
		go func() {
			if err := consumers.TaskConsumer.Consume(consumerCtx); err != nil {
				logger.Error("task consumer exited", zap.Error(err))
			}
		}()

		prevStop := stopMessaging
		stopMessaging = func() {
			consumerCancel()
			prevStop()
		}
	} else {
		stopMessaging = func() {}
		logger.Info("messaging disabled — consumers not started")
	}

	commandBus, authSvc, err := initializeCommands(
		persistenceLayer,
		externalServices,
		messagePublisher,
		cfg,
		redisClient,
		logger,
	)
	if err != nil {
		logger.Error("failed to initialize commands", zap.Error(err))
		return nil, err
	}

	// Started after the command bus exists, since the ticker dispatches onto it.
	stopScheduler := initializeScheduler(commandBus, cfg, logger)

	router := initializeHTTP(commandBus, logger, authSvc, redisClient, persistenceLayer.DB, cfg.Telemetry.Enabled, cfg)

	stop := func() {
		if router.RateLimiter != nil {
			router.RateLimiter.Stop()
		}

		// Before stopMessaging, so a tick already in flight can finish
		// publishing rather than losing its events.
		stopScheduler()

		stopMessaging()

		if tel != nil {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			tel.Shutdown(shutdownCtx)
		}

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

func (a *App) Close() {
	if a.DB != nil {
		if sqlDB, err := a.DB.DB(); err == nil {
			sqlDB.Close()
		}
	}
	if a.Redis != nil {
		a.Redis.Close()
	}
}
