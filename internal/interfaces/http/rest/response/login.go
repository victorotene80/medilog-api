package response

import "time"

type LoginResponse struct {
	UserID              string              `json:"user_id,omitempty"`
	Tokens              *AuthTokensResponse `json:"tokens,omitempty"`
	LastLogin           *time.Time          `json:"last_login,omitempty"`
	ChallengeID         *string             `json:"challenge_id,omitempty"`
	Status              string              `json:"status"`
	OnboardingCompleted bool                `json:"onboarding_completed"`
	RequiresOnboarding  bool                `json:"requires_onboarding"`
}

type GoogleAuthResponse struct {
	Status              string              `json:"status"`
	UserID              string              `json:"user_id,omitempty"`
	Email               string              `json:"email,omitempty"`
	FirstName           string              `json:"first_name,omitempty"`
	LastName            string              `json:"last_name,omitempty"`
	PictureURL          *string             `json:"picture_url,omitempty"`
	GoogleID            string              `json:"google_id,omitempty"`
	LastLogin           *time.Time          `json:"last_login,omitempty"`
	Tokens              *AuthTokensResponse `json:"tokens,omitempty"`
	OnboardingCompleted bool                `json:"onboarding_completed"`
	RequiresOnboarding  bool                `json:"requires_onboarding"`
}
