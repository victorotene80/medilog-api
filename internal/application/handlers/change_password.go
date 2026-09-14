package handlers

import (
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"

	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
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
	sessions        appContracts.SessionInvalidator
	auditLogger     appContracts.AuditLogger
	clock           func() time.Time
}

func NewChangePasswordHandler(
	userRepo repository.UserAggregateRepository,
	passwordHasher contracts.PasswordHasher,
	passwordService *domainServices.PasswordService,
	sessions appContracts.SessionInvalidator,
	auditLogger appContracts.AuditLogger,
	clock func() time.Time,
) *ChangePasswordHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &ChangePasswordHandler{
		userRepo:        userRepo,
		passwordHasher:  passwordHasher,
		passwordService: passwordService,
		sessions:        sessions,
		auditLogger:     auditLogger,
		clock:           clock,
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
		return struct{}{}, application.NewNotFound("user not found")
	}

	if agg.User.PasswordHash == nil || !h.passwordHasher.Verify(cmd.OldPassword, *agg.User.PasswordHash) {
		return struct{}{}, application.NewUnauthorized("current password is incorrect")
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

	// Force re-login everywhere. The version bump inside InvalidateAllForUser is
	// what evicts already-issued access tokens from the session cache.
	if err := h.sessions.InvalidateAllForUser(ctx, cmd.UserID, now); err != nil {
		return struct{}{}, fmt.Errorf("invalidate sessions: %w", err)
	}

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
