package response

import "time"

type GetUserResponse struct {
	ID                  string                      `json:"id"`
	Email               *string                     `json:"email,omitempty"`
	Phone               *string                     `json:"phone,omitempty"`
	FirstName           string                      `json:"first_name"`
	LastName            string                      `json:"last_name"`
	DOB                 *time.Time                  `json:"dob,omitempty"`
	Sex                 string                      `json:"sex,omitempty"`
	BloodType           string                      `json:"blood_type,omitempty"`
	Height              *float64                    `json:"height,omitempty"`
	Weight              *float64                    `json:"weight,omitempty"`
	WeightUnit          string                      `json:"weight_unit,omitempty"`
	TemperatureUnit     string                      `json:"temperature_unit,omitempty"`
	Country             string                      `json:"country,omitempty"`
	AvatarURL           *string                     `json:"avatar_url,omitempty"`
	CountryCode         string                      `json:"country_code,omitempty"`
	EmergencyContact    *GetUserEmergencyContactDTO `json:"emergency_contact,omitempty"`
	Status              string                      `json:"status"`
	OnboardingCompleted bool                        `json:"onboarding_completed"`
	LastLoginAt         *time.Time                  `json:"last_login_at,omitempty"`
	CreatedAt           time.Time                   `json:"created_at"`
	UpdatedAt           time.Time                   `json:"updated_at"`
}

type GetUserEmergencyContactDTO struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Relationship string `json:"relationship"`
	Phone        string `json:"phone"`
	IsPrimary    bool   `json:"is_primary"`
}
