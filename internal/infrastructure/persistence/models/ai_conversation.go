package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type AIConversationModel struct {
	ID                  int64      `gorm:"column:id;primaryKey;autoIncrement"`
	PublicID            string     `gorm:"column:public_id;type:uuid;default:gen_random_uuid()"`
	UserID              int64      `gorm:"column:user_id;not null;index"`
	Title               *string    `gorm:"column:title"`
	RelatedMedicationID *int64     `gorm:"column:related_medication_id"`
	RelatedVisitID      *int64     `gorm:"column:related_visit_id"`
	Summary             *string    `gorm:"column:summary"`
	SummaryUpTo         *int64     `gorm:"column:summary_up_to"`
	Status              string     `gorm:"column:status;not null;default:active"`
	LastMessageAt       *time.Time `gorm:"column:last_message_at"`
	CreatedAt           time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (AIConversationModel) TableName() string { return "ai_conversations" }

func AIConversationToEntity(m *AIConversationModel) (*entities.AIConversation, error) {
	status, err := valueobjects.NewConversationStatus(m.Status)
	if err != nil {
		return nil, err
	}
	return &entities.AIConversation{
		ID:                  m.ID,
		PublicID:            m.PublicID,
		UserID:              m.UserID,
		Title:               m.Title,
		RelatedMedicationID: m.RelatedMedicationID,
		RelatedVisitID:      m.RelatedVisitID,
		Summary:             m.Summary,
		SummaryUpTo:         m.SummaryUpTo,
		Status:              status,
		LastMessageAt:       m.LastMessageAt,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}, nil
}

func AIConversationToModel(e *entities.AIConversation) *AIConversationModel {
	return &AIConversationModel{
		ID:                  e.ID,
		PublicID:            e.PublicID,
		UserID:              e.UserID,
		Title:               e.Title,
		RelatedMedicationID: e.RelatedMedicationID,
		RelatedVisitID:      e.RelatedVisitID,
		Summary:             e.Summary,
		SummaryUpTo:         e.SummaryUpTo,
		Status:              e.Status.String(),
		LastMessageAt:       e.LastMessageAt,
		CreatedAt:           e.CreatedAt,
		UpdatedAt:           e.UpdatedAt,
	}
}
