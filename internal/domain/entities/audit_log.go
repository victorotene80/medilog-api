package entities

import "time"

type AuditLog struct {
	ID          int64
	Action      string
	UserID      *string
	ActorID     *string
	SessionID   *string
	IPAddress   *string
	UserAgent   *string
	CountryCode *string
	TargetID    *string
	Metadata    map[string]any
	Success     bool
	OccurredAt  time.Time
	CreatedAt   time.Time
}
