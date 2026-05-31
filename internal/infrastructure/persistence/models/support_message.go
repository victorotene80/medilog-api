package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type SupportMessageModel struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TicketID       int64     `gorm:"column:ticket_id;not null;index"`
	SenderUserID   *int64    `gorm:"column:sender_user_id"`
	Message        string    `gorm:"column:message;not null"`
	IsInternalNote bool      `gorm:"column:is_internal_note;not null;default:false"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (SupportMessageModel) TableName() string { return "support_messages" }

func SupportMessageToEntity(m *SupportMessageModel) *entities.SupportMessage {
	return &entities.SupportMessage{
		ID:             m.ID,
		TicketID:       m.TicketID,
		SenderUserID:   m.SenderUserID,
		Message:        m.Message,
		IsInternalNote: m.IsInternalNote,
		CreatedAt:      m.CreatedAt,
	}
}

func SupportMessageToModel(e *entities.SupportMessage) *SupportMessageModel {
	return &SupportMessageModel{
		ID:             e.ID,
		TicketID:       e.TicketID,
		SenderUserID:   e.SenderUserID,
		Message:        e.Message,
		IsInternalNote: e.IsInternalNote,
		CreatedAt:      e.CreatedAt,
	}
}
