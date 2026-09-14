package persistence

import (
	"context"
	"errors"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"

	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

var _ repository.SupportTicketRepository = (*SupportTicketRepository)(nil)

type SupportTicketRepository struct {
	db *gorm.DB
	// events drains the aggregate's domain events into the outbox on the same
	// transaction as the write. See drainAggregateEvents.
	events appContracts.MessagePublisher
}

func NewSupportTicketRepository(db *gorm.DB, events appContracts.MessagePublisher) *SupportTicketRepository {
	return &SupportTicketRepository{db: db, events: events}
}

func (r *SupportTicketRepository) FindByID(ctx context.Context, id int64) (*aggregates.SupportTicketAggregate, error) {
	var m models.SupportTicketModel
	if err := conn(ctx, r.db).Where("deleted_at IS NULL").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.hydrate(ctx, &m)
}

func (r *SupportTicketRepository) FindByUserID(ctx context.Context, userID int64) ([]*aggregates.SupportTicketAggregate, error) {
	var ms []models.SupportTicketModel
	if err := conn(ctx, r.db).Where("deleted_at IS NULL").Where("user_id = ?", userID).Order("created_at DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	if len(ms) == 0 {
		return nil, nil
	}

	ticketIDs := make([]int64, len(ms))
	for i, t := range ms {
		ticketIDs[i] = t.ID
	}

	var msgModels []models.SupportMessageModel
	if err := conn(ctx, r.db).Where("ticket_id IN ?", ticketIDs).Where("deleted_at IS NULL").Order("created_at ASC").Find(&msgModels).Error; err != nil {
		return nil, err
	}

	var attachModels []models.SupportAttachmentModel
	if len(msgModels) > 0 {
		msgIDs := make([]int64, len(msgModels))
		for i, mm := range msgModels {
			msgIDs[i] = mm.ID
		}
		if err := conn(ctx, r.db).Where("message_id IN ?", msgIDs).Where("deleted_at IS NULL").Find(&attachModels).Error; err != nil {
			return nil, err
		}
	}

	return r.hydrateAllWithData(ctx, ms, msgModels, attachModels)
}

func (r *SupportTicketRepository) FindOpenByUserID(ctx context.Context, userID int64) ([]*aggregates.SupportTicketAggregate, error) {
	var ms []models.SupportTicketModel
	if err := conn(ctx, r.db).
		Where("deleted_at IS NULL").
		Where("user_id = ? AND status IN ('open', 'in_progress')", userID).
		Order("created_at DESC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	if len(ms) == 0 {
		return nil, nil
	}

	ticketIDs := make([]int64, len(ms))
	for i, t := range ms {
		ticketIDs[i] = t.ID
	}

	var msgModels []models.SupportMessageModel
	if err := conn(ctx, r.db).Where("ticket_id IN ?", ticketIDs).Where("deleted_at IS NULL").Order("created_at ASC").Find(&msgModels).Error; err != nil {
		return nil, err
	}

	var attachModels []models.SupportAttachmentModel
	if len(msgModels) > 0 {
		msgIDs := make([]int64, len(msgModels))
		for i, mm := range msgModels {
			msgIDs[i] = mm.ID
		}
		if err := conn(ctx, r.db).Where("message_id IN ?", msgIDs).Where("deleted_at IS NULL").Find(&attachModels).Error; err != nil {
			return nil, err
		}
	}

	return r.hydrateAllWithData(ctx, ms, msgModels, attachModels)
}

func (r *SupportTicketRepository) Save(ctx context.Context, agg *aggregates.SupportTicketAggregate) error {
	return conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
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
		// Drained here, on the same transaction as the write: leaving it to the
		// caller meant most mutations raised events that were silently discarded.
		return drainAggregateEvents(ctx, tx, r.events, agg)
	})
}

func (r *SupportTicketRepository) Update(ctx context.Context, agg *aggregates.SupportTicketAggregate) error {
	if agg == nil || agg.Ticket == nil || agg.Ticket.ID <= 0 {
		return errors.New("support ticket id is required")
	}

	// Save inserts when the primary key is zero and reports success when it
	// matches nothing, so a ticket soft-deleted mid-request left the new message
	// orphaned under it with last_message_at never advanced — and a 200.
	// Wrapped in a transaction so the outbox envelope commits with the row it
	// describes; a bare Updates cannot carry the drain.
	return conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		result := tx.
			Model(&models.SupportTicketModel{}).
			Where("id = ?", agg.Ticket.ID).
			Select("*").
			Omit("id", "public_id", "created_at", "deleted_at").
			Updates(models.SupportTicketToModel(agg.Ticket))

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return repository.ErrNotFound
		}

		return drainAggregateEvents(ctx, tx, r.events, agg)
	})
}

func (r *SupportTicketRepository) hydrate(ctx context.Context, m *models.SupportTicketModel) (*aggregates.SupportTicketAggregate, error) {
	entity, err := models.SupportTicketToEntity(m)
	if err != nil {
		return nil, err
	}

	var msgModels []models.SupportMessageModel
	if err := conn(ctx, r.db).
		Where("ticket_id = ?", m.ID).
		Order("created_at ASC").
		Find(&msgModels).Error; err != nil {
		return nil, err
	}
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
		if err := conn(ctx, r.db).Where("message_id IN ?", msgIDs).Find(&attachModels).Error; err != nil {
			return nil, err
		}
	}
	attachments := make([]*entities.SupportAttachment, len(attachModels))
	for i, a := range attachModels {
		a := a
		attachments[i] = models.SupportAttachmentToEntity(&a)
	}

	return aggregates.RestoreSupportTicketAggregate(entity, msgs, attachments, 0), nil
}

func (r *SupportTicketRepository) hydrateAllWithData(ctx context.Context, ms []models.SupportTicketModel, allMsgModels []models.SupportMessageModel, allAttachModels []models.SupportAttachmentModel) ([]*aggregates.SupportTicketAggregate, error) {
	msgsByTicket := make(map[int64][]models.SupportMessageModel)
	for _, mm := range allMsgModels {
		msgsByTicket[mm.TicketID] = append(msgsByTicket[mm.TicketID], mm)
	}

	attachByMsg := make(map[int64][]models.SupportAttachmentModel)
	for _, a := range allAttachModels {
		attachByMsg[a.MessageID] = append(attachByMsg[a.MessageID], a)
	}

	result := make([]*aggregates.SupportTicketAggregate, 0, len(ms))
	for _, m := range ms {
		entity, err := models.SupportTicketToEntity(&m)
		if err != nil {
			return nil, err
		}

		ticketMsgs := msgsByTicket[m.ID]
		msgs := make([]*entities.SupportMessage, len(ticketMsgs))
		var msgIDs []int64
		for i, mm := range ticketMsgs {
			mm := mm
			msgs[i] = models.SupportMessageToEntity(&mm)
			msgIDs = append(msgIDs, mm.ID)
		}

		var attachments []*entities.SupportAttachment
		for _, msgID := range msgIDs {
			for _, a := range attachByMsg[msgID] {
				a := a
				attachments = append(attachments, models.SupportAttachmentToEntity(&a))
			}
		}

		result = append(result, aggregates.RestoreSupportTicketAggregate(entity, msgs, attachments, 0))
	}
	return result, nil
}
