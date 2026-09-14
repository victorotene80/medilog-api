package persistence

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type SupportAttachmentRepository struct {
	db *gorm.DB
}

func NewSupportAttachmentRepository(db *gorm.DB) *SupportAttachmentRepository {
	return &SupportAttachmentRepository{db: db}
}

func (r *SupportAttachmentRepository) FindByMessageID(ctx context.Context, messageID int64) ([]*entities.SupportAttachment, error) {
	var ms []models.SupportAttachmentModel
	if err := conn(ctx, r.db).Where("message_id = ?", messageID).Find(&ms).Error; err != nil {
		return nil, err
	}
	result := make([]*entities.SupportAttachment, len(ms))
	for i, m := range ms {
		m := m
		result[i] = models.SupportAttachmentToEntity(&m)
	}
	return result, nil
}

func (r *SupportAttachmentRepository) Save(ctx context.Context, attachment *entities.SupportAttachment) error {
	m := models.SupportAttachmentToModel(attachment)
	if err := conn(ctx, r.db).Create(m).Error; err != nil {
		return err
	}
	attachment.ID = m.ID
	return nil
}
