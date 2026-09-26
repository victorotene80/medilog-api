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

	// Links are never removed when an account is soft-deleted, so without the
	// join a deleted user's Google ID resolves to a user FindByID cannot load,
	// and that person can neither sign in nor sign up again.
	err := conn(ctx, r.db).
		Select("user_auth_providers.*").
		Joins("JOIN users ON users.id = user_auth_providers.user_id AND users.deleted_at IS NULL").
		Where("user_auth_providers.provider = ? AND user_auth_providers.provider_uid = ?", provider, providerUID).
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

	if err := conn(ctx, r.db).
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

	if err := conn(ctx, r.db).Create(model).Error; err != nil {
		return err
	}

	p.ID = model.ID
	return nil
}
