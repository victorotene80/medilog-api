package entities

import "time"

type Feedback struct {
	ID          int64
	PublicID    string
	UserID      *int64
	Rating      *int
	Title       *string
	Message     string
	AppVersion  *string
	Platform    *string
	DeviceModel *string
	Status      string
	CreatedAt   time.Time
}
