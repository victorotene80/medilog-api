package entities

import "time"

// AuditLog is a recorded security-relevant action.
//
// It exists so the query side has a domain type to return: the repository
// previously handed *models.AuditLogModel — a GORM struct with no json tags —
// all the way to the HTTP layer, which serialized it with Go field names as
// wire keys and shipped whatever columns the table happened to carry.
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
