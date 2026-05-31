package persistence

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type AIMessageRepository struct {
	db *gorm.DB
}

func NewAIMessageRepository(db *gorm.DB) *AIMessageRepository {
	return &AIMessageRepository{db: db}
}

func (r *AIMessageRepository) FindByConversationID(ctx context.Context, conversationID int64) ([]*entities.AIMessage, error) {
	var ms []models.AIMessageModel
	if err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return toAIMessageEntities(ms)
}

func (r *AIMessageRepository) FindByConversationIDSince(ctx context.Context, conversationID, afterMessageID int64) ([]*entities.AIMessage, error) {
	var ms []models.AIMessageModel
	if err := r.db.WithContext(ctx).
		Where("conversation_id = ? AND id > ?", conversationID, afterMessageID).
		Order("created_at ASC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return toAIMessageEntities(ms)
}

func (r *AIMessageRepository) Save(ctx context.Context, msg *entities.AIMessage) error {
	m := models.AIMessageToModel(msg)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	msg.ID = m.ID
	return nil
}

func (r *AIMessageRepository) SaveAll(ctx context.Context, msgs []*entities.AIMessage) error {
	ms := make([]models.AIMessageModel, len(msgs))
	for i, msg := range msgs {
		ms[i] = *models.AIMessageToModel(msg)
	}
	if err := r.db.WithContext(ctx).Create(&ms).Error; err != nil {
		return err
	}
	for i, m := range ms {
		msgs[i].ID = m.ID
	}
	return nil
}

func toAIMessageEntities(ms []models.AIMessageModel) ([]*entities.AIMessage, error) {
	result := make([]*entities.AIMessage, 0, len(ms))
	for _, m := range ms {
		m := m
		e, err := models.AIMessageToEntity(&m)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}
