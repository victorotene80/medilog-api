package handlers

import (
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"

	"context"
	"fmt"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	domainServices "github.com/victorotene80/medilog-api/internal/domain/services"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
	"github.com/victorotene80/medilog-api/internal/shared/requestmeta"
)

type ForgotPasswordHandler struct {
	userRepo    repository.UserRepository
	otpRepo     repository.OTPCodeRepository
	otpService  *domainServices.OTPService
	smsSender   appContracts.SMSSender
	auditLogger appContracts.AuditLogger
	clock       func() time.Time
	isLive      bool
}

func NewForgotPasswordHandler(
	userRepo repository.UserRepository,
	otpRepo repository.OTPCodeRepository,
	otpService *domainServices.OTPService,
	smsSender appContracts.SMSSender,
	auditLogger appContracts.AuditLogger,
	clock func() time.Time,
	isLive bool,
) *ForgotPasswordHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &ForgotPasswordHandler{
		userRepo:    userRepo,
		otpRepo:     otpRepo,
		otpService:  otpService,
		smsSender:   smsSender,
		auditLogger: auditLogger,
		clock:       clock,
		isLive:      isLive,
	}
}

func (h *ForgotPasswordHandler) Handle(
	ctx context.Context,
	cmd command.ForgotPasswordCommand,
) (struct{}, error) {
	now := h.clock()
	meta, _ := requestmeta.FromContext(ctx)

	agg, err := h.userRepo.FindByPhone(ctx, cmd.Recipient)
	if err != nil {
		return struct{}{}, nil
	}
	if agg == nil {
		agg, err = h.userRepo.FindByEmail(ctx, cmd.Recipient)
		if err != nil || agg == nil {
			return struct{}{}, nil
		}
	}

	if err := h.otpRepo.InvalidatePreviousByRecipientAndPurpose(
		ctx, cmd.Recipient, string(valueobjects.OTPPurposePasswordReset), now,
	); err != nil {
		return struct{}{}, fmt.Errorf("invalidate previous OTPs: %w", err)
	}

	channel := valueobjects.OTPChannelSMS
	if strings.Contains(cmd.Recipient, "@") {
		channel = valueobjects.OTPChannelEmail
	}

	otp, plainCode, err := h.otpService.NewOTP(
		agg.ID,
		cmd.Recipient,
		channel,
		valueobjects.OTPPurposePasswordReset,
	)
	if err != nil {
		return struct{}{}, fmt.Errorf("generate OTP: %w", err)
	}

	if !h.isLive {
		plainCode = "123456"
		hash, hashErr := h.otpService.Hash(plainCode)
		if hashErr != nil {
			return struct{}{}, fmt.Errorf("hash test OTP: %w", hashErr)
		}
		otp.CodeHash = hash
	}

	if err := h.otpRepo.Save(ctx, otp); err != nil {
		return struct{}{}, fmt.Errorf("save OTP: %w", err)
	}

	if h.isLive {
		message := fmt.Sprintf("Your MediLog password reset code is: %s. It expires in 10 minutes.", plainCode)
		if channel == valueobjects.OTPChannelEmail {
			if err := h.smsSender.Send(ctx, cmd.Recipient, message); err != nil {
				return struct{}{}, fmt.Errorf("send OTP email: %w", err)
			}
		} else {
			if err := h.smsSender.Send(ctx, cmd.Recipient, message); err != nil {
				return struct{}{}, fmt.Errorf("send OTP SMS: %w", err)
			}
		}
	}

	if h.auditLogger != nil {
		userID := fmt.Sprintf("%d", agg.ID)
		_ = h.auditLogger.Log(ctx, dto.AuditRecord{
			Action:     dto.AuditActionOTPRequested,
			UserID:     &userID,
			IPAddress:  &meta.IPAddress,
			UserAgent:  &meta.UserAgent,
			Success:    true,
			OccurredAt: now,
			Metadata: map[string]any{
				"purpose": valueobjects.OTPPurposePasswordReset,
				"channel": string(channel),
			},
		})
	}

	return struct{}{}, nil
}
