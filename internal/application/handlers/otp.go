package handlers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	domainServices "github.com/victorotene80/medilog-api/internal/domain/services"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type RequestOTPHandler struct {
	otpRepo    repository.OTPCodeRepository
	userRepo   repository.UserRepository
	otpService *domainServices.OTPService
	smsSender  appContracts.SMSSender
	clock      func() time.Time
}

func NewRequestOTPHandler(
	otpRepo repository.OTPCodeRepository,
	userRepo repository.UserRepository,
	otpService *domainServices.OTPService,
	smsSender appContracts.SMSSender,
	clock func() time.Time,
) *RequestOTPHandler {
	if clock == nil {
		clock = func() time.Time {
			return time.Now().UTC()
		}
	}

	return &RequestOTPHandler{
		otpRepo:    otpRepo,
		userRepo:   userRepo,
		otpService: otpService,
		smsSender:  smsSender,
		clock:      clock,
	}
}

func (h *RequestOTPHandler) Handle(
	ctx context.Context,
	cmd command.RequestOTPCommand,
) (*dto.RequestOTPResultDTO, error) {
	recipient := strings.TrimSpace(cmd.Recipient)
	if recipient == "" {
		return nil, errors.New("recipient is required")
	}

	userID, err := h.resolveOTPUserID(ctx, recipient)
	if err != nil {
		return nil, err
	}

	channel, err := valueobjects.NewOTPChannel(strings.TrimSpace(cmd.Channel))
	if err != nil {
		return nil, err
	}

	purpose, err := valueobjects.NewOTPPurpose(strings.TrimSpace(cmd.Purpose))
	if err != nil {
		return nil, err
	}

	now := h.clock().UTC()

	if err := h.otpRepo.InvalidatePreviousByRecipientAndPurpose(
		ctx,
		recipient,
		purpose.String(),
		now,
	); err != nil {
		return nil, fmt.Errorf("invalidate previous OTP: %w", err)
	}

	otp, plainCode, err := h.otpService.NewOTP(
		userID,
		recipient,
		channel,
		purpose,
	)
	if err != nil {
		return nil, fmt.Errorf("generate OTP: %w", err)
	}

	if err := h.otpRepo.Save(ctx, otp); err != nil {
		return nil, fmt.Errorf("save OTP: %w", err)
	}

	expiresInMinutes := int(otp.ExpiresAt.Sub(now).Minutes())
	if expiresInMinutes < 1 {
		expiresInMinutes = 1
	}

	message := buildOTPMessage(
		purpose.String(),
		plainCode,
		expiresInMinutes,
	)

	if channel == valueobjects.OTPChannelSMS {
		if h.smsSender == nil {
			return nil, errors.New("sms sender is required")
		}

		if err := h.smsSender.Send(ctx, recipient, message); err != nil {
			return nil, fmt.Errorf("send sms OTP: %w", err)
		}
	}

	return &dto.RequestOTPResultDTO{
		Recipient: recipient,
		Channel:   channel.String(),
		Purpose:   purpose.String(),
		ExpiresAt: otp.ExpiresAt,
		Message:   "OTP sent successfully",
	}, nil
}

type VerifyOTPHandler struct {
	otpRepo    repository.OTPCodeRepository
	otpService *domainServices.OTPService
	clock      func() time.Time
}

func NewVerifyOTPHandler(
	otpRepo repository.OTPCodeRepository,
	otpService *domainServices.OTPService,
	clock func() time.Time,
) *VerifyOTPHandler {
	if clock == nil {
		clock = func() time.Time {
			return time.Now().UTC()
		}
	}

	return &VerifyOTPHandler{
		otpRepo:    otpRepo,
		otpService: otpService,
		clock:      clock,
	}
}

func (h *VerifyOTPHandler) Handle(
	ctx context.Context,
	cmd command.VerifyOTPCommand,
) (*dto.VerifyOTPResultDTO, error) {
	recipient := strings.TrimSpace(cmd.Recipient)
	if recipient == "" {
		return nil, errors.New("recipient is required")
	}

	code := strings.TrimSpace(cmd.Code)
	if code == "" {
		return nil, errors.New("OTP code is required")
	}

	channel, err := valueobjects.NewOTPChannel(strings.TrimSpace(cmd.Channel))
	if err != nil {
		return nil, err
	}

	purpose, err := valueobjects.NewOTPPurpose(strings.TrimSpace(cmd.Purpose))
	if err != nil {
		return nil, err
	}

	otp, err := h.otpRepo.FindLatestByRecipientAndPurpose(
		ctx,
		recipient,
		purpose.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("find OTP: %w", err)
	}

	if otp == nil {
		return nil, errors.New("invalid or expired OTP")
	}

	now := h.clock().UTC()

	if !otp.IsValid(now) {
		return nil, errors.New("invalid or expired OTP")
	}

	if otp.Channel != channel.String() {
		return nil, errors.New("invalid or expired OTP")
	}

	if !h.otpService.Verify(code, otp.CodeHash) {
		return nil, errors.New("invalid or expired OTP")
	}

	otp.MarkUsed(now)

	if err := h.otpRepo.Update(ctx, otp); err != nil {
		return nil, fmt.Errorf("mark OTP used: %w", err)
	}

	return &dto.VerifyOTPResultDTO{
		UserID:    strconv.FormatInt(otp.UserID, 10),
		Recipient: recipient,
		Channel:   channel.String(),
		Purpose:   purpose.String(),
		Verified:  true,
		Message:   "OTP verified successfully",
	}, nil
}

func buildOTPMessage(
	purpose string,
	code string,
	expiresInMinutes int,
) string {
	if expiresInMinutes < 1 {
		expiresInMinutes = 1
	}

	switch purpose {
	case "verify_onboarding":
		return fmt.Sprintf(
			"Your Medilog onboarding verification code is %s. It expires in %d minutes. Do not share this code with anyone.",
			code,
			expiresInMinutes,
		)

	case "login":
		return fmt.Sprintf(
			"Your Medilog login code is %s. It expires in %d minutes. Do not share this code with anyone.",
			code,
			expiresInMinutes,
		)

	case "password_reset":
		return fmt.Sprintf(
			"Your Medilog password reset code is %s. It expires in %d minutes. Do not share this code with anyone.",
			code,
			expiresInMinutes,
		)

	case "phone_verification":
		return fmt.Sprintf(
			"Your Medilog phone verification code is %s. It expires in %d minutes. Do not share this code with anyone.",
			code,
			expiresInMinutes,
		)

	default:
		return fmt.Sprintf(
			"Your Medilog verification code is %s. It expires in %d minutes. Do not share this code with anyone.",
			code,
			expiresInMinutes,
		)
	}
}

func (h *RequestOTPHandler) resolveOTPUserID(
	ctx context.Context,
	recipient string,
) (int64, error) {
	recipient = strings.TrimSpace(recipient)
	if recipient == "" {
		return 0, errors.New("recipient is required")
	}

	if strings.Contains(recipient, "@") {
		user, err := h.userRepo.FindByEmail(ctx, recipient)
		if err != nil {
			return 0, fmt.Errorf("find user by email: %w", err)
		}

		if user == nil {
			return 0, errors.New("user not found")
		}

		return user.ID, nil
	}

	user, err := h.userRepo.FindByPhone(ctx, recipient)
	if err != nil {
		return 0, fmt.Errorf("find user by phone: %w", err)
	}

	if user == nil {
		return 0, errors.New("user not found")
	}

	return user.ID, nil
}
