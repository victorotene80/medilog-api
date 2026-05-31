package entities

import "time"

type SupportMessage struct {
	ID             int64
	TicketID       int64
	SenderUserID   *int64 // nil = system message
	Message        string
	IsInternalNote bool
	CreatedAt      time.Time
}

func (m *SupportMessage) IsSystemMessage() bool {
	return m.SenderUserID == nil
}
