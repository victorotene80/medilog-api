package handlers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/contracts"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	domainServices "github.com/victorotene80/medilog-api/internal/domain/services"
	"github.com/victorotene80/medilog-api/internal/shared/requestmeta"
)

type ChangePasswordHandler struct {
	userRepo        repository.UserAggregateRepository
	passwordHasher  contracts.PasswordHasher
	passwordService *domainServices.PasswordService
	refreshRepo     repository.RefreshTokenRepository
	sessionCache    appContracts.Cache[string, appContracts.CachedToken]
	auditLogger     appContracts.AuditLogger
	clock           func() time.Time
	version         appContracts.SessionCache
}

func NewChangePasswordHandler(
	userRepo repository.UserAggregateRepository,
	passwordHasher contracts.PasswordHasher,
	passwordService *domainServices.PasswordService,
	refreshRepo repository.RefreshTokenRepository,
	sessionCache appContracts.Cache[string, appContracts.CachedToken],
	auditLogger appContracts.AuditLogger,
	version appContracts.SessionCache,
	clock func() time.Time,
) *ChangePasswordHandler {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}
	return &ChangePasswordHandler{
		userRepo:        userRepo,
		passwordHasher:  passwordHasher,
		passwordService: passwordService,
		refreshRepo:     refreshRepo,
		sessionCache:    sessionCache,
		auditLogger:     auditLogger,
		clock:           clock,
		version:         version,
	}
}

func (h *ChangePasswordHandler) Handle(
	ctx context.Context,
	cmd command.ChangePasswordCommand,
) (struct{}, error) {
	now := h.clock()
	meta, _ := requestmeta.FromContext(ctx)

	agg, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil || agg == nil {
		return struct{}{}, errors.New("user not found")
	}

	if agg.User.PasswordHash == nil || !h.passwordHasher.Verify(cmd.OldPassword, *agg.User.PasswordHash) {
		return struct{}{}, errors.New("current password is incorrect")
	}

	if err := h.passwordService.Validate(cmd.NewPassword); err != nil {
		return struct{}{}, err
	}

	newHash, err := h.passwordHasher.Hash(cmd.NewPassword)
	if err != nil {
		return struct{}{}, fmt.Errorf("hash new password: %w", err)
	}

	agg.ChangePassword(newHash, now)

	if err := h.userRepo.Update(ctx, agg); err != nil {
		return struct{}{}, fmt.Errorf("save updated password: %w", err)
	}

	// revoke all sessions — force re-login everywhere
	if err := h.refreshRepo.RevokeAllForUser(ctx, cmd.UserID, now); err != nil {
		return struct{}{}, fmt.Errorf("revoke sessions: %w", err)
	}

	if h.version != nil {
		_ = h.version.IncrementVersion(ctx, cmd.UserID)
	}
	// best-effort cache invalidation — we don't have all session IDs here
	// so we rely on TTL expiry for stale cache entries
	// if you want immediate invalidation, store a per-user token version in cache

	if h.auditLogger != nil {
		userID := cmd.UserID
		userIDText := strconv.FormatInt(userID, 10)

		_ = h.auditLogger.Log(ctx, dto.AuditRecord{
			Action:     dto.AuditActionPasswordChanged,
			UserID:     &userIDText,
			ActorID:    &userIDText,
			IPAddress:  &meta.IPAddress,
			UserAgent:  &meta.UserAgent,
			Success:    true,
			OccurredAt: now,
		})
	}

	return struct{}{}, nil
}
