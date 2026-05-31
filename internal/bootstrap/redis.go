package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/victorotene80/medilog-api/internal/shared/config"
)

func initializeRedis(cfg config.RedisConfig, logger *zap.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		if logger != nil {
			logger.Error("redis ping failed",
				zap.String("addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)),
				zap.Int("db", cfg.DB),
				zap.Error(err),
			)
		}
		return nil, fmt.Errorf("redis init failed: %w", err)
	}

	if logger != nil {
		logger.Info("redis connected",
			zap.String("addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)),
			zap.Int("db", cfg.DB),
		)
	}

	return client, nil
}