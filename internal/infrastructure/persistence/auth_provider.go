package persistence

import (
	"context"
	"errors"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

var _ repository.UserAuthProviderRepository = (*UserAuthProviderRepository)(nil)

type UserAuthProviderRepository struct {
	db *gorm.DB
}

func NewUserAuthProviderRepository(db *gorm.DB) (*UserAuthProviderRepository, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	return &UserAuthProviderRepository{db: db}, nil
}

func (r *UserAuthProviderRepository) FindByProviderUID(
	ctx context.Context,
	provider string,
	providerUID string,
) (*entities.UserAuthProvider, error) {
	var model models.UserAuthProviderModel

	err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_uid = ?", provider, providerUID).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return models.UserAuthProviderModelToEntity(&model), nil
}

func (r *UserAuthProviderRepository) FindByUserID(
	ctx context.Context,
	userID int64,
) ([]*entities.UserAuthProvider, error) {
	var modelList []models.UserAuthProviderModel

	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&modelList).Error; err != nil {
		return nil, err
	}

	result := make([]*entities.UserAuthProvider, len(modelList))
	for i := range modelList {
		result[i] = models.UserAuthProviderModelToEntity(&modelList[i])
	}
	return result, nil
}

func (r *UserAuthProviderRepository) Create(
	ctx context.Context,
	p *entities.UserAuthProvider,
) error {
	if p == nil {
		return errors.New("auth provider is required")
	}

	model := models.UserAuthProviderEntityToModel(p)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	p.ID = model.ID
	return nil
}
