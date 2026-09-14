package handlers

import (
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"

	"context"
	"fmt"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	domainServices "github.com/victorotene80/medilog-api/internal/domain/services"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type DeleteAccountHandler struct {
	userRepo    repository.UserAggregateRepository
	otpRepo     repository.OTPCodeRepository
	sessions    appContracts.SessionInvalidator
	otpService  *domainServices.OTPService
	auditLogger appContracts.AuditLogger
	clock       func() time.Time
}

func NewDeleteAccountHandler(
	userRepo repository.UserAggregateRepository,
	otpRepo repository.OTPCodeRepository,
	sessions appContracts.SessionInvalidator,
	otpService *domainServices.OTPService,
	auditLogger appContracts.AuditLogger,
	clock func() time.Time,
) *DeleteAccountHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &DeleteAccountHandler{
		userRepo:    userRepo,
		otpRepo:     otpRepo,
		sessions:    sessions,
		otpService:  otpService,
		auditLogger: auditLogger,
		clock:       clock,
	}
}

func (h *DeleteAccountHandler) Handle(
	ctx context.Context,
	cmd command.DeleteAccountCommand,
) (struct{}, error) {
	now := h.clock()

	otp, err := h.otpRepo.FindLatestByRecipientAndPurpose(ctx, cmd.Recipient, string(valueobjects.OTPPurposeDeleteAccount))
	if err != nil {
		return struct{}{}, fmt.Errorf("find OTP: %w", err)
	}

	if otp == nil {
		return struct{}{}, application.NewUnauthorized("invalid or expired OTP")
	}

	// cmd.Recipient is caller-supplied, so without this an OTP legitimately issued
	// for one account would authorize deleting a different one.
	if otp.UserID != cmd.UserID {
		return struct{}{}, application.NewUnauthorized("invalid or expired OTP")
	}

	// Gate before verifying, and gate on IsValid rather than re-deriving it: it
	// covers expiry, single-use *and* the per-code attempt allowance. Checking
	// these after Verify meant an expired or spent code was still compared, and
	// an attempt-exhausted one was never rejected at all.
	if !otp.IsValid(now) {
		return struct{}{}, application.NewUnauthorized("invalid or expired OTP")
	}

	// Verify takes the plaintext code; passing a hash here compares a digest
	// against the string of another digest and never matches.
	if !h.otpService.Verify(cmd.OTPCode, otp.CodeHash) {
		recordFailedOTPAttempt(ctx, h.otpRepo, otp)
		return struct{}{}, application.NewUnauthorized("invalid OTP code")
	}

	otp.UsedAt = &now
	if err := h.otpRepo.Update(ctx, otp); err != nil {
		return struct{}{}, fmt.Errorf("mark OTP used: %w", err)
	}

	agg, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return struct{}{}, fmt.Errorf("find user: %w", err)
	}

	if agg == nil {
		return struct{}{}, application.NewNotFound("user not found")
	}

	if err := agg.SoftDelete(now); err != nil {
		return struct{}{}, fmt.Errorf("soft delete user: %w", err)
	}

	if err := h.userRepo.Update(ctx, agg); err != nil {
		return struct{}{}, fmt.Errorf("update user: %w", err)
	}

	// Ordered after the soft delete on purpose: these are destructive but
	// recoverable, so a failure here leaves a deleted account rather than a live
	// account whose owner has been signed out of every device.
	if err := h.sessions.InvalidateAllForUser(ctx, cmd.UserID, now); err != nil {
		return struct{}{}, fmt.Errorf("invalidate sessions: %w", err)
	}

	if h.auditLogger != nil {
		userIDStr := fmt.Sprintf("%d", cmd.UserID)
		// Best-effort: the account is already deleted, so a failed audit write must
		// not fail the command. Matches logout.go.
		_ = h.auditLogger.Log(ctx, dto.AuditRecord{
			Action:     dto.AuditAction("account.deleted"),
			UserID:     &userIDStr,
			ActorID:    &userIDStr,
			Success:    true,
			OccurredAt: now,
		})
	}

	return struct{}{}, nil
}
