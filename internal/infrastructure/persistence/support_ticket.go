package persistence

import (
	"context"
	"errors"

	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type SupportTicketRepository struct {
	db *gorm.DB
}

func NewSupportTicketRepository(db *gorm.DB) *SupportTicketRepository {
	return &SupportTicketRepository{db: db}
}

func (r *SupportTicketRepository) FindByID(ctx context.Context, id int64) (*aggregates.SupportTicketAggregate, error) {
	var m models.SupportTicketModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.hydrate(ctx, &m)
}

func (r *SupportTicketRepository) FindByUserID(ctx context.Context, userID int64) ([]*aggregates.SupportTicketAggregate, error) {
	var ms []models.SupportTicketModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	return r.hydrateAll(ctx, ms)
}

func (r *SupportTicketRepository) FindOpenByUserID(ctx context.Context, userID int64) ([]*aggregates.SupportTicketAggregate, error) {
	var ms []models.SupportTicketModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND status IN ('open', 'in_progress')", userID).
		Order("created_at DESC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return r.hydrateAll(ctx, ms)
}

func (r *SupportTicketRepository) Save(ctx context.Context, agg *aggregates.SupportTicketAggregate) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ticketModel := models.SupportTicketToModel(agg.Ticket)
		if err := tx.Create(ticketModel).Error; err != nil {
			return err
		}
		agg.Ticket.ID = ticketModel.ID
		agg.Ticket.PublicID = ticketModel.PublicID

		for _, msg := range agg.Messages {
			msg.TicketID = ticketModel.ID
			msgModel := models.SupportMessageToModel(msg)
			if err := tx.Create(msgModel).Error; err != nil {
				return err
			}
			msg.ID = msgModel.ID
		}
		return nil
	})
}

func (r *SupportTicketRepository) Update(ctx context.Context, agg *aggregates.SupportTicketAggregate) error {
	return r.db.WithContext(ctx).Save(models.SupportTicketToModel(agg.Ticket)).Error
}

func (r *SupportTicketRepository) hydrate(ctx context.Context, m *models.SupportTicketModel) (*aggregates.SupportTicketAggregate, error) {
	entity, err := models.SupportTicketToEntity(m)
	if err != nil {
		return nil, err
	}

	var msgModels []models.SupportMessageModel
	r.db.WithContext(ctx).Where("ticket_id = ?", m.ID).Order("created_at ASC").Find(&msgModels)
	msgs := make([]*entities.SupportMessage, len(msgModels))
	for i, mm := range msgModels {
		mm := mm
		msgs[i] = models.SupportMessageToEntity(&mm)
	}

	var attachModels []models.SupportAttachmentModel
	if len(msgModels) > 0 {
		msgIDs := make([]int64, len(msgModels))
		for i, mm := range msgModels {
			msgIDs[i] = mm.ID
		}
		r.db.WithContext(ctx).Where("message_id IN ?", msgIDs).Find(&attachModels)
	}
	attachments := make([]*entities.SupportAttachment, len(attachModels))
	for i, a := range attachModels {
		a := a
		attachments[i] = models.SupportAttachmentToEntity(&a)
	}

	return aggregates.RestoreSupportTicketAggregate(entity, msgs, attachments, 0), nil
}

func (r *SupportTicketRepository) hydrateAll(ctx context.Context, ms []models.SupportTicketModel) ([]*aggregates.SupportTicketAggregate, error) {
	result := make([]*aggregates.SupportTicketAggregate, 0, len(ms))
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
