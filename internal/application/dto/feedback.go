package dto

import "time"

type FeedbackResponseDTO struct {
	ID          string    `json:"id"`
	UserID      *string   `json:"user_id,omitempty"`
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
