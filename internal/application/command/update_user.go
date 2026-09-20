package command

import "time"

// UpdateUserCommand is a partial update of the caller's own profile. Every
// field is a pointer: nil means "leave unchanged".
//
// Email and phone are deliberately absent. Changing either has to re-run
// verification, so they keep their existing OTP-backed flow rather than being
// silently writable here.
type UpdateUserCommand struct {
	UserID int64

	FirstName   *string
	LastName    *string
	DOB         *time.Time
	Sex         *int
	BloodType   *string
	AvatarURL   *string
	CountryCode *string

	Height          *float64
	Weight          *float64
	WeightUnit      *string
	TemperatureUnit *string
	Timezone        *string
}
