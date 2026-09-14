package dto

import "time"

type GetUserDTO struct {
	ID        string
	Email     *string
	Phone     *string
	FirstName string
	LastName  string
	DOB       *time.Time
	Sex       string
	BloodType string
	Height    *float64
	Weight    *float64
	// Units accompany the readings. Without them a client renders a bare "72"
	// and cannot tell kg from lb — the profile stores the unit and lets the user
	// change it, so the number alone is ambiguous.
	WeightUnit          string
	TemperatureUnit     string
	Country             string
	AvatarURL           *string
	CountryCode         string
	EmergencyContact    *GetUserEmergencyContactDTO
	Status              string
	OnboardingCompleted bool
	LastLoginAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type GetUserEmergencyContactDTO struct {
	ID           string
	Name         string
	Relationship string
	Phone        string
	CountryCode  string
	IsPrimary    bool
}
