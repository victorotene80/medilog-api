package dto

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/contracts"
)

type GoogleAuthStatus string

const (
	GoogleAuthStatusLogin    GoogleAuthStatus = "LOGIN"
	GoogleAuthStatusRegister GoogleAuthStatus = "REGISTER"
)

type GoogleAuthResultDTO struct {
	Status              GoogleAuthStatus
	UserID              string
	Email               string
	FirstName           string
	LastName            string
	PictureURL          *string
	GoogleID            string
	LastLogin           *time.Time
	Tokens              *contracts.TokenPair
	OnboardingCompleted bool
	RequiresOnboarding  bool
}
