package dto

import "time"

// FeedbackResponseDTO deliberately carries no user identifier: feedback may be
// submitted anonymously, and echoing an owner back to a reader served no client
// purpose while leaking who filed what.
type FeedbackResponseDTO struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Rating      *int      `json:"rating,omitempty"`
	Title       *string   `json:"title,omitempty"`
	Message     string    `json:"message"`
	AppVersion  *string   `json:"app_version,omitempty"`
	Platform    *string   `json:"platform,omitempty"`
	DeviceModel *string   `json:"device_model,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
