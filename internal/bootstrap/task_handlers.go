package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application/messaging"
	coremsg "github.com/victorotene80/medilog-api/internal/infrastructure/messaging"
	"go.uber.org/zap"
)

type SendWelcomeEmailPayload struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
}

type SendVerificationPayload struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Code   string `json:"code"`
}

type SyncAnalyticsUserPayload struct {
	UserID int64  `json:"user_id"`
	Action string `json:"action"`
}

func HandleSendWelcomeEmail(logger *zap.Logger) coremsg.HandlerFunc {
	return func(ctx context.Context, env messaging.Envelope) error {
		var payload SendWelcomeEmailPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal send-welcome-email payload: %w", err)
		}

		logger.Info("sending welcome email",
			zap.Int64("user_id", payload.UserID),
			zap.String("email", payload.Email),
			zap.String("task_id", env.ID),
		)

		return nil
	}
}

func HandleSendVerification(logger *zap.Logger) coremsg.HandlerFunc {
	return func(ctx context.Context, env messaging.Envelope) error {
		var payload SendVerificationPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal send-verification payload: %w", err)
		}

		logger.Info("sending verification",
			zap.Int64("user_id", payload.UserID),
			zap.String("email", payload.Email),
			zap.String("task_id", env.ID),
		)

		return nil
	}
}

func HandleSyncAnalyticsUser(logger *zap.Logger) coremsg.HandlerFunc {
	return func(ctx context.Context, env messaging.Envelope) error {
		var payload SyncAnalyticsUserPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal sync-analytics-user payload: %w", err)
		}

		logger.Info("syncing analytics user",
			zap.Int64("user_id", payload.UserID),
			zap.String("action", payload.Action),
			zap.String("task_id", env.ID),
		)

		return nil
	}
}
