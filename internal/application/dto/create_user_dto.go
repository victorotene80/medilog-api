package dto

import "time"

type RegisterUserResponseDTO struct {
	UserID    string
	Email     string
	FirstName string
	LastName  string
	Phone     string
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time

	OnboardingCompleted bool
	RequiresOnboarding  bool
}
