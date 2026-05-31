package persistence

import (
	"context"
	"errors"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type UserProfileRepository struct {
	db *gorm.DB
}

func NewUserProfileRepository(db *gorm.DB) *UserProfileRepository {
	return &UserProfileRepository{db: db}
}

func (r *UserProfileRepository) FindByUserID(ctx context.Context, userID int64) (*entities.UserProfile, error) {
	var m models.UserProfileModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.UserProfileModelToEntity(&m), nil
}

func (r *UserProfileRepository) Save(ctx context.Context, profile *entities.UserProfile) error {
	m := models.UserProfileEntityToModel(profile)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	profile.ID = m.ID
	return nil
}

func (r *UserProfileRepository) Update(ctx context.Context, profile *entities.UserProfile) error {
	return r.db.WithContext(ctx).Save(models.UserProfileEntityToModel(profile)).Error
}
