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
	"github.com/victorotene80/medilog-api/internal/domain/contracts"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	domainServices "github.com/victorotene80/medilog-api/internal/domain/services"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
	"github.com/victorotene80/medilog-api/internal/shared/requestmeta"
)

type VerifyOnboardingOTPHandler struct {
	userRepository repository.UserRepository
	otpRepo        repository.OTPCodeRepository
	otpService     *domainServices.OTPService
	sessionService appContracts.SessionService
	clock          func() time.Time
}

func NewVerifyOnboardingOTPHandler(
	userRepository repository.UserRepository,
	otpRepo repository.OTPCodeRepository,
	otpService *domainServices.OTPService,
	sessionService appContracts.SessionService,
	clock func() time.Time,
) *VerifyOnboardingOTPHandler {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}

	return &VerifyOnboardingOTPHandler{
		userRepository: userRepository,
		otpRepo:        otpRepo,
		otpService:     otpService,
		sessionService: sessionService,
		clock:          clock,
	}
}

func (h *VerifyOnboardingOTPHandler) Handle(
	ctx context.Context,
	cmd command.VerifyOnboardingOTPCommand,
) (*dto.VerifyOnboardingOTPResultDTO, error) {

	recipient := strings.TrimSpace(cmd.Recipient)
	if recipient == "" {
		return nil, errors.New("recipient is required")
	}

	channel, err := valueobjects.NewOTPChannel(strings.TrimSpace(cmd.Channel))
	if err != nil {
		return nil, err
	}

	purpose, err := valueobjects.NewOTPPurpose(strings.TrimSpace(cmd.Purpose))
	if err != nil {
		return nil, err
	}

	code := strings.TrimSpace(cmd.Code)
	if code == "" {
		return nil, errors.New("code is required")
	}

	if purpose != valueobjects.OTPPurposeEmailVerification &&
		purpose != valueobjects.OTPPurposePhoneVerification {
		return nil, errors.New("invalid onboarding OTP purpose")
	}

	otp, err := h.otpRepo.FindLatestByRecipientAndPurpose(ctx, recipient, purpose.String())
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

	if otp.UserID <= 0 {
		return nil, errors.New("OTP is not linked to a user")
	}

	user, err := h.userRepository.FindByID(ctx, otp.UserID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	switch purpose {
	case valueobjects.OTPPurposeEmailVerification:
		user.EmailVerifiedAt = &now

	case valueobjects.OTPPurposePhoneVerification:
		user.PhoneVerifiedAt = &now
	}

	if user.Status.IsPendingVerification() {
		user.Status = valueobjects.UserStatusActive
	}

	user.UpdatedAt = now

	otp.MarkUsed(now)

	if err := h.otpRepo.Update(ctx, otp); err != nil {
		return nil, fmt.Errorf("mark OTP used: %w", err)
	}

	if err := h.userRepository.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user verification: %w", err)
	}

	meta, _ := requestmeta.FromContext(ctx)

	sessionResult, err := h.sessionService.Create(
		ctx,
		strconv.FormatInt(user.ID, 10),
		meta.IPAddress,
		meta.UserAgent,
		meta.DeviceID,
		meta.DeviceFingerprint,
		meta.DeviceName,
	)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &dto.VerifyOnboardingOTPResultDTO{
		UserID:    user.PublicID,
		Recipient: recipient,
		Channel:   channel.String(),
		Purpose:   purpose.String(),
		Verified:  true,
		Message:   "OTP verified successfully",

		Tokens: contracts.TokenPair{
			AccessToken: contracts.Token{
				Value:     sessionResult.AccessToken.Value,
				ExpiresAt: sessionResult.AccessToken.ExpiresAt,
			},
			RefreshToken: contracts.Token{
				Value:     sessionResult.RefreshToken.Value,
				ExpiresAt: sessionResult.RefreshToken.ExpiresAt,
			},
		},

		OnboardingCompleted: user.IsOnboardingCompleted,
		RequiresOnboarding:  !user.IsOnboardingCompleted,
		OnboardingStep:      onboardingStep(user.IsOnboardingCompleted),

		AccessTokenExpiresAt:  sessionResult.AccessToken.ExpiresAt,
		RefreshTokenExpiresAt: sessionResult.RefreshToken.ExpiresAt,
	}, nil
}

func onboardingStep(completed bool) string {
	if completed {
		return "completed"
	}

	return "emergency_contact"
}
