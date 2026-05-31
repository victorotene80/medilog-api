package entities

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type AIConversation struct {
	ID                  int64
	PublicID            string
	UserID              int64
	Title               *string
	RelatedMedicationID *int64
	RelatedVisitID      *int64
	Summary             *string // rolling compressed context
	SummaryUpTo         *int64  // last message ID included in summary
	Status              valueobjects.ConversationStatus
	LastMessageAt       *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (c *AIConversation) IsActive() bool {
	return c.Status == valueobjects.ConversationStatusActive
}

func (c *AIConversation) Archive(now time.Time) {
	c.Status = valueobjects.ConversationStatusArchived
	c.UpdatedAt = now
}

func (c *AIConversation) UpdateSummary(summary string, upToMessageID int64, now time.Time) {
	c.Summary = &summary
	c.SummaryUpTo = &upToMessageID
	c.UpdatedAt = now
}

func (c *AIConversation) TouchLastMessage(now time.Time) {
	c.LastMessageAt = &now
	c.UpdatedAt = now
}
