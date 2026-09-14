package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application/messaging"
	coremsg "github.com/victorotene80/medilog-api/internal/infrastructure/messaging"
	"go.uber.org/zap"
)

type AuthUserCreatedPayload struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type AuthSessionPayload struct {
	UserID    int64  `json:"user_id"`
	SessionID string `json:"session_id"`
}

type AuthPasswordChangedPayload struct {
	UserID int64 `json:"user_id"`
}

func HandleAuthUserCreated(logger *zap.Logger) coremsg.HandlerFunc {
	return func(ctx context.Context, env messaging.Envelope) error {
		var payload AuthUserCreatedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal user.created payload: %w", err)
		}

		logger.Info("user.created event received",
			zap.Int64("user_id", payload.UserID),
			zap.String("email", payload.Email),
			zap.String("event_id", env.ID),
		)

		return nil
	}
}

func HandleAuthUserLocked(logger *zap.Logger) coremsg.HandlerFunc {
	return func(ctx context.Context, env messaging.Envelope) error {
		var payload struct {
			UserID int64 `json:"user_id"`
		}
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal user.locked payload: %w", err)
		}

		logger.Info("user.locked event received",
			zap.Int64("user_id", payload.UserID),
			zap.String("event_id", env.ID),
		)

		return nil
	}
}

func HandleAuthSessionCreated(logger *zap.Logger) coremsg.HandlerFunc {
	return func(ctx context.Context, env messaging.Envelope) error {
		var payload AuthSessionPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal session.created payload: %w", err)
		}

		logger.Info("session.created event received",
			zap.Int64("user_id", payload.UserID),
			zap.String("session_id", payload.SessionID),
			zap.String("event_id", env.ID),
		)

		return nil
	}
}

func HandleAuthSessionRevoked(logger *zap.Logger) coremsg.HandlerFunc {
	return func(ctx context.Context, env messaging.Envelope) error {
		var payload AuthSessionPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal session.revoked payload: %w", err)
		}

		logger.Info("session.revoked event received",
			zap.Int64("user_id", payload.UserID),
			zap.String("session_id", payload.SessionID),
			zap.String("event_id", env.ID),
		)

		return nil
	}
}

func HandleAuthPasswordChanged(logger *zap.Logger) coremsg.HandlerFunc {
	return func(ctx context.Context, env messaging.Envelope) error {
		var payload AuthPasswordChangedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal password.changed payload: %w", err)
		}

		logger.Info("password.changed event received",
			zap.Int64("user_id", payload.UserID),
			zap.String("event_id", env.ID),
		)

		return nil
	}
}
