package dto

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/contracts"
)

type LoginResultDTO struct {
	UserID string

	Tokens contracts.TokenPair

	LastLogin   *time.Time
	ChallengeID *string
	Status      string

	OnboardingCompleted bool
	RequiresOnboarding  bool
}
