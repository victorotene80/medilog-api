package persistence

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type SupportMessageRepository struct {
	db *gorm.DB
}

func NewSupportMessageRepository(db *gorm.DB) *SupportMessageRepository {
	return &SupportMessageRepository{db: db}
}

func (r *SupportMessageRepository) FindByTicketID(ctx context.Context, ticketID int64) ([]*entities.SupportMessage, error) {
	var ms []models.SupportMessageModel
	if err := r.db.WithContext(ctx).Where("ticket_id = ?", ticketID).Order("created_at ASC").Find(&ms).Error; err != nil {
		return nil, err
	}
	result := make([]*entities.SupportMessage, len(ms))
	for i, m := range ms {
		m := m
		result[i] = models.SupportMessageToEntity(&m)
	}
	return result, nil
}

func (r *SupportMessageRepository) Save(ctx context.Context, msg *entities.SupportMessage) error {
	m := models.SupportMessageToModel(msg)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	msg.ID = m.ID
	return nil
}
