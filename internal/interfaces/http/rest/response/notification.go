package response

import "time"

type NotificationResponse struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	Title       string         `json:"title"`
	Body        string         `json:"body"`
	Type        string         `json:"type"`
	Channel     *string        `json:"channel,omitempty"`
	Status      string         `json:"status"`
	ImageURL    *string        `json:"image_url,omitempty"`
	ScheduledAt *time.Time     `json:"scheduled_at,omitempty"`
	SentAt      *time.Time     `json:"sent_at,omitempty"`
	ReadAt      *time.Time     `json:"read_at,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type ListNotificationsResponse struct {
	Notifications []*NotificationResponse `json:"notifications"`
}
