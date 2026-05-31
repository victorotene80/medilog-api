package persistence

import (
	"context"
	"errors"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type FeedbackRepository struct {
	db *gorm.DB
}

func NewFeedbackRepository(db *gorm.DB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

func (r *FeedbackRepository) FindByID(ctx context.Context, id int64) (*entities.Feedback, error) {
	var m models.FeedbackModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.FeedbackToEntity(&m), nil
}

func (r *FeedbackRepository) Save(ctx context.Context, f *entities.Feedback) error {
	m := models.FeedbackToModel(f)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	f.ID = m.ID
	return nil
}

func (r *FeedbackRepository) Update(ctx context.Context, f *entities.Feedback) error {
	return r.db.WithContext(ctx).Save(models.FeedbackToModel(f)).Error
}
