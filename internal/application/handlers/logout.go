package handlers

import (
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"

	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/shared/requestmeta"
)

type LogoutHandler struct {
	refreshRepo  repository.RefreshTokenRepository
	sessionCache appContracts.Cache[string, appContracts.CachedToken]
	auditLogger  appContracts.AuditLogger
	clock        func() time.Time
}

func NewLogoutHandler(
	refreshRepo repository.RefreshTokenRepository,
	sessionCache appContracts.Cache[string, appContracts.CachedToken],
	auditLogger appContracts.AuditLogger,
	clock func() time.Time,
) *LogoutHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &LogoutHandler{
		refreshRepo:  refreshRepo,
		sessionCache: sessionCache,
		auditLogger:  auditLogger,
		clock:        clock,
	}
}

func (h *LogoutHandler) Handle(ctx context.Context, cmd command.LogoutCommand) (struct{}, error) {
	now := h.clock()
	meta, _ := requestmeta.FromContext(ctx)

	rtID, err := strconv.ParseInt(cmd.SessionID, 10, 64)
	if err != nil {
		return struct{}{}, fmt.Errorf("invalid session ID")
	}

	rt, err := h.refreshRepo.FindByID(ctx, rtID)
	if err != nil || rt == nil {
		return struct{}{}, nil // idempotent — already gone is fine
	}

	if !rt.IsRevoked() {
		rt.Revoke(now, nil)
		if err := h.refreshRepo.Update(ctx, rt); err != nil {
			return struct{}{}, fmt.Errorf("revoke session: %w", err)
		}
	}

	if h.sessionCache != nil {
		_ = h.sessionCache.Delete(ctx, cmd.SessionID)
	}

	if h.auditLogger != nil {
		userID := cmd.UserID
		userIDText := strconv.FormatInt(userID, 10)
		sessionID := cmd.SessionID
		_ = h.auditLogger.Log(ctx, dto.AuditRecord{
			Action:     dto.AuditActionLogout,
			UserID:     &userIDText,
			ActorID:    &userIDText,
			SessionID:  &sessionID,
			IPAddress:  &meta.IPAddress,
			UserAgent:  &meta.UserAgent,
			Success:    true,
			OccurredAt: now,
		})
	}

	return struct{}{}, nil
}
