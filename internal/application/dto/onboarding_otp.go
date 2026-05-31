package dto

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/contracts"
)

type VerifyOnboardingOTPResultDTO struct {
	UserID string

	Recipient string
	Channel   string
	Purpose   string
	Verified  bool
	Message   string

	Tokens contracts.TokenPair

	OnboardingCompleted bool
	RequiresOnboarding  bool
	OnboardingStep      string

	AccessTokenExpiresAt  time.Time
	RefreshTokenExpiresAt time.Time
}
