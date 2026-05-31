package entities

import "time"

type EmergencyContact struct {
	ID           int64
	PublicID     string
	UserID       int64
	Name         string
	Relationship string
	Phone        string
	IsPrimary    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
