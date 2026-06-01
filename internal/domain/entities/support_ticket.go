package entities

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type SupportAttachment struct {
	ID        int64
	PublicID  string
	MessageID int64
	FileURL   string
	FileName  *string
	FileType  *string
	FileSize  *int64
	CreatedAt time.Time
}

type SupportTicket struct {
	ID            int64
	PublicID      string
	UserID        int64
	CategoryID    int
	Status        valueobjects.TicketStatus
	Priority      valueobjects.TicketPriority
	AssignedTo    *int64
	LastMessageAt *time.Time
	Description   *string
	CreatedAt     time.Time
	ClosedAt      *time.Time
	DateAssigned  *time.Time
}

func (t *SupportTicket) IsOpen() bool {
	return t.Status == valueobjects.TicketStatusOpen ||
		t.Status == valueobjects.TicketStatusInProgress
}

func (t *SupportTicket) Close(now time.Time) {
	t.Status = valueobjects.TicketStatusClosed
	t.ClosedAt = &now
}

func (t *SupportTicket) Assign(agentID int64, now time.Time) {
	t.AssignedTo = &agentID
	t.DateAssigned = &now
	t.Status = valueobjects.TicketStatusInProgress
}
