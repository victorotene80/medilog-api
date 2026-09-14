package response

import "time"

type EmergencyContactResponse struct {
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

// ListEmergencyContactsResponse is the payload for GET /emergency-contacts.
type ListEmergencyContactsResponse struct {
	Contacts []EmergencyContactResponse `json:"contacts"`
}
