package dto

type CreateUserDTO struct {
	UserID              int64
	Email               *string
	Phone               *string
	FirstName           string
	LastName            string
	OnboardingCompleted bool
	RequiresOnboarding  bool
}
