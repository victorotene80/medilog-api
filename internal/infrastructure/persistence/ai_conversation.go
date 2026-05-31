package persistence

import (
	"context"
	"errors"

	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type AIConversationRepository struct {
	db *gorm.DB
}

func NewAIConversationRepository(db *gorm.DB) *AIConversationRepository {
	return &AIConversationRepository{db: db}
}

func (r *AIConversationRepository) FindByID(ctx context.Context, id int64) (*aggregates.AIConversationAggregate, error) {
	var m models.AIConversationModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.hydrate(ctx, &m)
}

func (r *AIConversationRepository) FindByPublicID(ctx context.Context, userID int64, publicID string) (*aggregates.AIConversationAggregate, error) {
	var m models.AIConversationModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND public_id = ?", userID, publicID).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.hydrate(ctx, &m)
}

func (r *AIConversationRepository) FindByUserID(ctx context.Context, userID int64) ([]*aggregates.AIConversationAggregate, error) {
	var ms []models.AIConversationModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("last_message_at DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	return r.hydrateAll(ctx, ms)
}

func (r *AIConversationRepository) FindActiveByUserID(ctx context.Context, userID int64) ([]*aggregates.AIConversationAggregate, error) {
	var ms []models.AIConversationModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = 'active'", userID).
		Order("last_message_at DESC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return r.hydrateAll(ctx, ms)
}

func (r *AIConversationRepository) Save(ctx context.Context, agg *aggregates.AIConversationAggregate) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		convModel := models.AIConversationToModel(agg.Conversation)
		if err := tx.Create(convModel).Error; err != nil {
			return err
		}
		agg.Conversation.ID = convModel.ID
		agg.Conversation.PublicID = convModel.PublicID

		for _, msg := range agg.Messages {
			msg.ConversationID = convModel.ID
			msgModel := models.AIMessageToModel(msg)
			if err := tx.Create(msgModel).Error; err != nil {
				return err
			}
			msg.ID = msgModel.ID
		}
		return nil
	})
}

func (r *AIConversationRepository) Update(ctx context.Context, agg *aggregates.AIConversationAggregate) error {
	return r.db.WithContext(ctx).Save(models.AIConversationToModel(agg.Conversation)).Error
}

func (r *AIConversationRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&models.AIConversationModel{}).
		Where("id = ?", id).
		UpdateColumn("status", "deleted").Error
}

func (r *AIConversationRepository) hydrate(ctx context.Context, m *models.AIConversationModel) (*aggregates.AIConversationAggregate, error) {
	entity, err := models.AIConversationToEntity(m)
	if err != nil {
		return nil, err
	}

	var msgModels []models.AIMessageModel
	r.db.WithContext(ctx).
		Where("conversation_id = ?", m.ID).
		Order("created_at ASC").
		Find(&msgModels)

	msgs := make([]*entities.AIMessage, 0, len(msgModels))
	for _, mm := range msgModels {
		mm := mm
		msg, err := models.AIMessageToEntity(&mm)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}

	return aggregates.RestoreAIConversationAggregate(entity, msgs, 0), nil
}

func (r *AIConversationRepository) hydrateAll(ctx context.Context, ms []models.AIConversationModel) ([]*aggregates.AIConversationAggregate, error) {
	result := make([]*aggregates.AIConversationAggregate, 0, len(ms))
	for _, m := range ms {
		m := m
		agg, err := r.hydrate(ctx, &m)
		if err != nil {
			return nil, err
		}
		result = append(result, agg)
	}
	return result, nil
}
