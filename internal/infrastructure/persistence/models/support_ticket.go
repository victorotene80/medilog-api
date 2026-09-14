package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
	"gorm.io/gorm"
)

type SupportTicketModel struct {
	ID            int64          `gorm:"column:id;primaryKey;autoIncrement"`
	PublicID      string         `gorm:"column:public_id;type:uuid;default:gen_random_uuid()"`
	UserID        int64          `gorm:"column:user_id;not null;index"`
	CategoryID    int            `gorm:"column:category_id;not null"`
	Status        string         `gorm:"column:status;not null;default:open;index"`
	Priority      string         `gorm:"column:priority;not null;default:normal"`
	AssignedTo    *int64         `gorm:"column:assigned_to"`
	LastMessageAt *time.Time     `gorm:"column:last_message_at"`
	Description   *string        `gorm:"column:description"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime"`
	ClosedAt      *time.Time     `gorm:"column:closed_at"`
	DateAssigned  *time.Time     `gorm:"column:date_assigned"`
}

func (SupportTicketModel) TableName() string { return "support_tickets" }

func SupportTicketToEntity(m *SupportTicketModel) (*entities.SupportTicket, error) {
	status, err := valueobjects.NewTicketStatus(m.Status)
	if err != nil {
		return nil, err
	}
	priority, err := valueobjects.NewTicketPriority(m.Priority)
	if err != nil {
		return nil, err
	}
	return &entities.SupportTicket{
		ID:            m.ID,
		PublicID:      m.PublicID,
		UserID:        m.UserID,
		CategoryID:    m.CategoryID,
		Status:        status,
		Priority:      priority,
		AssignedTo:    m.AssignedTo,
		LastMessageAt: m.LastMessageAt,
		Description:   m.Description,
		CreatedAt:     m.CreatedAt,
		ClosedAt:      m.ClosedAt,
		DateAssigned:  m.DateAssigned,
	}, nil
}

func SupportTicketToModel(e *entities.SupportTicket) *SupportTicketModel {
	return &SupportTicketModel{
		ID:            e.ID,
		PublicID:      e.PublicID,
		UserID:        e.UserID,
		CategoryID:    e.CategoryID,
		Status:        e.Status.String(),
		Priority:      e.Priority.String(),
		AssignedTo:    e.AssignedTo,
		LastMessageAt: e.LastMessageAt,
		Description:   e.Description,
		CreatedAt:     e.CreatedAt,
		ClosedAt:      e.ClosedAt,
		DateAssigned:  e.DateAssigned,
	}
}
