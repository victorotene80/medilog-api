package response

type CreateUserResponse struct {
	UserID              string  `json:"user_id"`
	Email               *string `json:"email,omitempty"`
	Phone               *string `json:"phone,omitempty"`
	FirstName           string  `json:"first_name"`
	LastName            string  `json:"last_name"`
	OnboardingCompleted bool    `json:"onboarding_completed"`
	RequiresOnboarding  bool    `json:"requires_onboarding"`
}
