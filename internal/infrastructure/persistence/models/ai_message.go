package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type AIMessageModel struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ConversationID int64     `gorm:"column:conversation_id;not null;index:idx_ai_messages_conversation_created,priority:1"`
	Role           string    `gorm:"column:role;not null"`
	Content        string    `gorm:"column:content;not null"`
	TokenCount     *int      `gorm:"column:token_count"`
	Meta           JSONMap   `gorm:"column:meta;type:jsonb"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime;index:idx_ai_messages_conversation_created,priority:2"`
}

func (AIMessageModel) TableName() string { return "ai_messages" }

func AIMessageToEntity(m *AIMessageModel) (*entities.AIMessage, error) {
	role, err := valueobjects.NewAIRole(m.Role)
	if err != nil {
		return nil, err
	}
	return &entities.AIMessage{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		Role:           role,
		Content:        m.Content,
		TokenCount:     m.TokenCount,
		Meta:           map[string]any(m.Meta),
		CreatedAt:      m.CreatedAt,
	}, nil
}

func AIMessageToModel(e *entities.AIMessage) *AIMessageModel {
	return &AIMessageModel{
		ID:             e.ID,
		ConversationID: e.ConversationID,
		Role:           e.Role.String(),
		Content:        e.Content,
		TokenCount:     e.TokenCount,
		Meta:           JSONMap(e.Meta),
		CreatedAt:      e.CreatedAt,
	}
}
