package handlers

import (
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"

	"context"
	"fmt"
	"strconv"
	"strings"
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

type LoginViaOTPHandler struct {
	userRepo            repository.UserAggregateRepository
	otpRepo             repository.OTPCodeRepository
	otpService          *domainServices.OTPService
	sessionService      appContracts.SessionService
	sessionVersionCache appContracts.SessionCache
	auditLogger         appContracts.AuditLogger
	eventPublisher      appContracts.MessagePublisher
	clock               func() time.Time
}

func NewLoginViaOTPHandler(
	userRepo repository.UserAggregateRepository,
	otpRepo repository.OTPCodeRepository,
	otpService *domainServices.OTPService,
	sessionService appContracts.SessionService,
	sessionVersionCache appContracts.SessionCache,
	auditLogger appContracts.AuditLogger,
	eventPublisher appContracts.MessagePublisher,
	clock func() time.Time,
) *LoginViaOTPHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &LoginViaOTPHandler{
		userRepo:            userRepo,
		otpRepo:             otpRepo,
		otpService:          otpService,
		sessionService:      sessionService,
		sessionVersionCache: sessionVersionCache,
		auditLogger:         auditLogger,
		eventPublisher:      eventPublisher,
		clock:               clock,
	}
}

func (h *LoginViaOTPHandler) Handle(
	ctx context.Context,
	cmd command.LoginViaOTPCommand,
) (*contracts.TokenPair, error) {
	now := h.clock()
	meta, _ := requestmeta.FromContext(ctx)

	recipient := strings.TrimSpace(cmd.Recipient)
	if recipient == "" {
		return nil, application.NewValidation("recipient is required")
	}

	code := strings.TrimSpace(cmd.Code)
	if code == "" {
		return nil, application.NewValidation("OTP code is required")
	}

	channel, err := valueobjects.NewOTPChannel(strings.TrimSpace(cmd.Channel))
	if err != nil {
		return nil, application.NewValidation("invalid channel")
	}

	otp, err := h.otpRepo.FindLatestByRecipientAndPurpose(ctx, recipient, "login")
	if err != nil {
		// A database failure is not a bad code. Reporting it as one tells a user
		// with a valid OTP that their code is wrong, hides the outage from 5xx
		// dashboards, and stops client back-off from engaging.
		return nil, fmt.Errorf("find OTP: %w", err)
	}

	if otp == nil {
		return nil, application.NewUnauthorized("invalid or expired OTP")
	}

	if !otp.IsValid(now) {
		return nil, application.NewUnauthorized("invalid or expired OTP")
	}

	if otp.Channel != channel.String() {
		return nil, application.NewUnauthorized("invalid or expired OTP")
	}

	if !h.otpService.Verify(code, otp.CodeHash) {
		recordFailedOTPAttempt(ctx, h.otpRepo, otp)
		return nil, application.NewUnauthorized("invalid OTP code")
	}

	otp.MarkUsed(now)
	if err := h.otpRepo.Update(ctx, otp); err != nil {
		return nil, fmt.Errorf("mark OTP used: %w", err)
	}

	user, err := h.userRepo.FindByID(ctx, otp.UserID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	if user == nil {
		return nil, application.NewNotFound("user not found")
	}

	if user.IsLocked(now) {
		return nil, application.NewForbidden("account is locked")
	}

	user.RecordSuccessfulLogin(meta.IPAddress, now)
	if err := h.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	// Published by UserAggregateRepository.Update above, inside its transaction.

	sessionResult, err := h.sessionService.Create(
		ctx,
		fmt.Sprintf("%d", user.User.ID),
		meta.IPAddress,
		meta.UserAgent,
		cmd.DeviceID,
		"",
		cmd.DeviceName,
	)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	if h.auditLogger != nil {
		userIDText := strconv.FormatInt(user.User.ID, 10)
		sessionID := sessionResult.SessionID
		// Passwordless login is the highest-value account-takeover path in this
		// API and was the only auth flow leaving no audit trail.
		_ = h.auditLogger.Log(ctx, dto.AuditRecord{
			Action:     dto.AuditActionLoginSuccess,
			UserID:     &userIDText,
			ActorID:    &userIDText,
			SessionID:  &sessionID,
			IPAddress:  &meta.IPAddress,
			UserAgent:  &meta.UserAgent,
			Success:    true,
			OccurredAt: now,
		})
	}

	return &contracts.TokenPair{
		AccessToken: contracts.Token{
			Value:     sessionResult.AccessToken.Value,
			ExpiresAt: sessionResult.AccessToken.ExpiresAt,
		},
		RefreshToken: contracts.Token{
			Value:     sessionResult.RefreshToken.Value,
			ExpiresAt: sessionResult.RefreshToken.ExpiresAt,
		},
	}, nil
}
