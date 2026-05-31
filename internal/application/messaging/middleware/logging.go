package middleware

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/shared/requestmeta"
)

func Logging(logger *zap.Logger) messaging.Middleware {
	if logger == nil {
		logger = zap.NewNop()
	}

	return func(next messaging.HandlerFunc) messaging.HandlerFunc {
		return func(ctx context.Context, cmd messaging.Command) (any, error) {
			start := time.Now()
			cmdName := fmt.Sprintf("%T", cmd)

			meta, _ := requestmeta.FromContext(ctx)

			logger.Info("command started",
				zap.String("command", cmdName),
				zap.String("request_id", meta.RequestID),
				zap.String("ip_address", meta.IPAddress),
			)

			res, err := next(ctx, cmd)

			elapsed := time.Since(start)

			if err != nil {
				logger.Error("command failed",
					zap.String("command", cmdName),
					zap.Duration("duration", elapsed),
					zap.Error(err),
				)
			} else {
				logger.Info("command finished",
					zap.String("command", cmdName),
					zap.Duration("duration", elapsed),
				)
			}

			return res, err
		}
	}
}

func AttachLogging(bus *messaging.CommandBus, logger *zap.Logger) {
	if bus == nil {
		return
	}
	bus.Use(Logging(logger))
}
