package entities

import "time"

type Notification struct {
	ID          int64
	PublicID    string
	UserID      int64
	Title       string
	Body        string
	Type        string
	Channel     *string
	Status      string
	ImageURL    *string
	ScheduledAt *time.Time
	SentAt      *time.Time
	ReadAt      *time.Time
	Metadata    map[string]any
	// DedupeKey identifies a server-generated notification so the scheduler can
	// re-scan an overlapping window without ever emitting a duplicate. nil for
	// hand-authored notifications, which are not deduplicated.
	DedupeKey *string
	CreatedAt time.Time
}

func (n *Notification) IsRead() bool { return n.ReadAt != nil }
func (n *Notification) IsSent() bool { return n.SentAt != nil }

func (n *Notification) MarkRead(now time.Time) {
	n.ReadAt = &now
	n.Status = "read"
}
