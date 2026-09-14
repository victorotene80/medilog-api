package dto

import "time"

// EmergencyContactDTO carries an emergency contact across the application
// boundary. ID is the public UUID, never the internal row id.
type EmergencyContactDTO struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	Relationship string    `json:"relationship"`
	Phone        string    `json:"phone"`
	CountryCode  string    `json:"country_code"`
	IsPrimary    bool      `json:"is_primary"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
