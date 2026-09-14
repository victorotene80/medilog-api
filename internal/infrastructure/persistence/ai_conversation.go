package persistence

import (
	"context"
	"errors"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

var _ repository.AIConversationRepository = (*AIConversationRepository)(nil)

type AIConversationRepository struct {
	db *gorm.DB
	// events drains the aggregate's domain events into the outbox on the same
	// transaction as the write. See drainAggregateEvents.
	events appContracts.MessagePublisher
}

func NewAIConversationRepository(db *gorm.DB, events appContracts.MessagePublisher) *AIConversationRepository {
	return &AIConversationRepository{db: db, events: events}
}

func (r *AIConversationRepository) FindByID(ctx context.Context, id int64) (*aggregates.AIConversationAggregate, error) {
	var m models.AIConversationModel
	if err := conn(ctx, r.db).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.hydrate(ctx, &m)
}

func (r *AIConversationRepository) FindByPublicID(ctx context.Context, userID int64, publicID string) (*aggregates.AIConversationAggregate, error) {
	var m models.AIConversationModel
	if err := conn(ctx, r.db).
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
	if err := conn(ctx, r.db).Where("user_id = ?", userID).Order("last_message_at DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	if len(ms) == 0 {
		return nil, nil
	}

	convIDs := make([]int64, len(ms))
	for i, m := range ms {
		convIDs[i] = m.ID
	}

	var allMsgModels []models.AIMessageModel
	if err := conn(ctx, r.db).Where("conversation_id IN ?", convIDs).Order("created_at ASC").Find(&allMsgModels).Error; err != nil {
		return nil, err
	}

	return r.hydrateAllWithData(ctx, ms, allMsgModels)
}

func (r *AIConversationRepository) FindActiveByUserID(ctx context.Context, userID int64) ([]*aggregates.AIConversationAggregate, error) {
	var ms []models.AIConversationModel
	if err := conn(ctx, r.db).
		Where("user_id = ? AND status = 'active'", userID).
		Order("last_message_at DESC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return r.hydrateAll(ctx, ms)
}

func (r *AIConversationRepository) Save(ctx context.Context, agg *aggregates.AIConversationAggregate) error {
	return conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
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
		// Drained here, on the same transaction as the write: leaving it to the
		// caller meant most mutations raised events that were silently discarded.
		return drainAggregateEvents(ctx, tx, r.events, agg)
	})
}

func (r *AIConversationRepository) Update(ctx context.Context, agg *aggregates.AIConversationAggregate) error {
	if agg == nil || agg.Conversation == nil {
		return errors.New("ai conversation is required")
	}
	if agg.Conversation.ID <= 0 {
		return errors.New("ai conversation id is required")
	}

	model := models.AIConversationToModel(agg.Conversation)
	// Wrapped in a transaction so the outbox envelope commits with the row it
	// describes; a bare Updates cannot carry the drain.
	return conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		result := tx.
			Model(&models.AIConversationModel{}).
			Where("id = ?", agg.Conversation.ID).
			Select("*").
			Omit("id", "public_id", "created_at", "deleted_at").
			Updates(model)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return repository.ErrNotFound
		}

		return drainAggregateEvents(ctx, tx, r.events, agg)
	})
}

func (r *AIConversationRepository) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("ai conversation id is required")
	}

	result := conn(ctx, r.db).Model(&models.AIConversationModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     "deleted",
			"deleted_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *AIConversationRepository) hydrate(ctx context.Context, m *models.AIConversationModel) (*aggregates.AIConversationAggregate, error) {
	entity, err := models.AIConversationToEntity(m)
	if err != nil {
		return nil, err
	}

	var msgModels []models.AIMessageModel
	// Checked: an empty history here is indistinguishable from a first turn, so
	// the model would answer a follow-up medical question with no context and no
	// error anywhere.
	if err := conn(ctx, r.db).
		Where("conversation_id = ?", m.ID).
		Order("created_at ASC").
		Find(&msgModels).Error; err != nil {
		return nil, err
	}

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

func (r *AIConversationRepository) hydrateAllWithData(ctx context.Context, ms []models.AIConversationModel, allMsgModels []models.AIMessageModel) ([]*aggregates.AIConversationAggregate, error) {
	msgsByConv := make(map[int64][]models.AIMessageModel)
	for _, mm := range allMsgModels {
		msgsByConv[mm.ConversationID] = append(msgsByConv[mm.ConversationID], mm)
	}

	result := make([]*aggregates.AIConversationAggregate, 0, len(ms))
	for _, m := range ms {
		entity, err := models.AIConversationToEntity(&m)
		if err != nil {
			return nil, err
		}

		convMsgs := msgsByConv[m.ID]
		msgs := make([]*entities.AIMessage, 0, len(convMsgs))
		for _, mm := range convMsgs {
			mm := mm
			msg, err := models.AIMessageToEntity(&mm)
			if err != nil {
				return nil, err
			}
			msgs = append(msgs, msg)
		}

		result = append(result, aggregates.RestoreAIConversationAggregate(entity, msgs, 0))
	}
	return result, nil
}
