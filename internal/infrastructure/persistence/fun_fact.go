package persistence

import (
	"context"
	"errors"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type FunFactRepository struct {
	db *gorm.DB
}

func NewFunFactRepository(db *gorm.DB) *FunFactRepository {
	return &FunFactRepository{db: db}
}

func (r *FunFactRepository) FindByID(ctx context.Context, id int64) (*entities.FunFact, error) {
	var m models.FunFactModel

	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return models.FunFactToEntity(&m), nil
}

func (r *FunFactRepository) FindAll(
	ctx context.Context,
	activeOnly bool,
) ([]*entities.FunFact, error) {
	query := r.db.WithContext(ctx).Model(&models.FunFactModel{})
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	var ms []models.FunFactModel
	if err := query.Order("created_at DESC").Find(&ms).Error; err != nil {
		return nil, err
	}

	result := make([]*entities.FunFact, len(ms))
	for i, m := range ms {
		m := m
		result[i] = models.FunFactToEntity(&m)
	}
	return result, nil
}

func (r *FunFactRepository) Save(ctx context.Context, fact *entities.FunFact) error {
	if fact == nil {
		return errors.New("fun fact is required")
	}

	m := models.FunFactToModel(fact)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}

	fact.ID = m.ID
	fact.CreatedAt = m.CreatedAt
	return nil
}

func (r *FunFactRepository) Update(ctx context.Context, fact *entities.FunFact) error {
	if fact == nil {
		return errors.New("fun fact is required")
	}
	if fact.ID <= 0 {
		return errors.New("fun fact id is required")
	}

	m := models.FunFactToModel(fact)

	result := r.db.WithContext(ctx).
		Model(&models.FunFactModel{}).
		Where("id = ?", fact.ID).
		Select("*").
		Omit("id", "created_at", "deleted_at").
		Updates(m)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *FunFactRepository) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("fun fact id is required")
	}

	result := r.db.WithContext(ctx).Delete(&models.FunFactModel{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
