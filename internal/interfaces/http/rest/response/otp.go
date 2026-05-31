package response

import "time"

type RequestOTPResponse struct {
	Recipient string    `json:"recipient"`
	Channel   string    `json:"channel"`
	Purpose   string    `json:"purpose"`
	ExpiresAt time.Time `json:"expires_at"`
	Message   string    `json:"message"`
}

type VerifyOTPResponse struct {
	UserID    string `json:"user_id"`
	Recipient string `json:"recipient"`
	Channel   string `json:"channel"`
	Purpose   string `json:"purpose"`
	Verified  bool   `json:"verified"`
	Message   string `json:"message"`
}

type VerifyOnboardingOTPResponse struct {
	UserID string `json:"user_id"`

	Recipient string `json:"recipient"`
	Channel   string `json:"channel"`
	Purpose   string `json:"purpose"`
	Verified  bool   `json:"verified"`
	Message   string `json:"message"`

	Tokens *AuthTokensResponse `json:"tokens,omitempty"`

	OnboardingCompleted bool   `json:"onboarding_completed"`
	RequiresOnboarding  bool   `json:"requires_onboarding"`
	OnboardingStep      string `json:"onboarding_step"`
}
