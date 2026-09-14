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
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
	"github.com/victorotene80/medilog-api/internal/shared/requestmeta"
)

type ResetPasswordHandler struct {
	userRepo        repository.UserAggregateRepository
	otpRepo         repository.OTPCodeRepository
	sessions        appContracts.SessionInvalidator
	passwordHasher  contracts.PasswordHasher
	passwordService *domainServices.PasswordService
	otpService      *domainServices.OTPService
	auditLogger     appContracts.AuditLogger
	clock           func() time.Time
	isLive          bool
}

func NewResetPasswordHandler(
	userRepo repository.UserAggregateRepository,
	otpRepo repository.OTPCodeRepository,
	sessions appContracts.SessionInvalidator,
	passwordHasher contracts.PasswordHasher,
	passwordService *domainServices.PasswordService,
	otpService *domainServices.OTPService,
	auditLogger appContracts.AuditLogger,
	clock func() time.Time,
	isLive bool,
) *ResetPasswordHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &ResetPasswordHandler{
		userRepo:        userRepo,
		otpRepo:         otpRepo,
		sessions:        sessions,
		passwordHasher:  passwordHasher,
		passwordService: passwordService,
		otpService:      otpService,
		auditLogger:     auditLogger,
		clock:           clock,
		isLive:          isLive,
	}
}

func (h *ResetPasswordHandler) Handle(
	ctx context.Context,
	cmd command.ResetPasswordCommand,
) (struct{}, error) {
	now := h.clock()
	meta, _ := requestmeta.FromContext(ctx)

	otp, err := h.otpRepo.FindLatestByRecipientAndPurpose(
		ctx, cmd.Recipient, string(valueobjects.OTPPurposePasswordReset),
	)
	if err != nil || otp == nil {
		return struct{}{}, application.NewValidation("invalid or expired reset code")
	}

	bypass := !h.isLive && cmd.OTPCode == "123456"

	if !bypass {
		if !otp.IsValid(now) {
			return struct{}{}, application.NewValidation("invalid or expired reset code")
		}

		if !h.otpService.Verify(cmd.OTPCode, otp.CodeHash) {
			recordFailedOTPAttempt(ctx, h.otpRepo, otp)
			return struct{}{}, application.NewValidation("invalid or expired reset code")
		}
	}

	if otp.UserID <= 0 {
		return struct{}{}, application.NewValidation("invalid reset code")
	}

	agg, err := h.userRepo.FindByID(ctx, otp.UserID)
	if err != nil || agg == nil {
		return struct{}{}, application.NewNotFound("user not found")
	}

	if err := h.passwordService.Validate(cmd.NewPassword); err != nil {
		return struct{}{}, err
	}

	newHash, err := h.passwordHasher.Hash(cmd.NewPassword)
	if err != nil {
		return struct{}{}, fmt.Errorf("hash password: %w", err)
	}

	agg.ChangePassword(newHash, now)

	if err := h.userRepo.Update(ctx, agg); err != nil {
		return struct{}{}, fmt.Errorf("save user: %w", err)
	}

	otp.MarkUsed(now)
	if err := h.otpRepo.Update(ctx, otp); err != nil {
		return struct{}{}, fmt.Errorf("mark OTP used: %w", err)
	}

	if err := h.sessions.InvalidateAllForUser(ctx, otp.UserID, now); err != nil {
		return struct{}{}, fmt.Errorf("invalidate sessions: %w", err)
	}

	if h.auditLogger != nil {
		userID := strconv.FormatInt(otp.UserID, 10)
		_ = h.auditLogger.Log(ctx, dto.AuditRecord{
			Action:     dto.AuditActionPasswordReset,
			UserID:     &userID,
			IPAddress:  &meta.IPAddress,
			UserAgent:  &meta.UserAgent,
			Success:    true,
			OccurredAt: now,
		})
	}

	return struct{}{}, nil
}
