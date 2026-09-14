package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

var _ repository.AllergyRepository = (*AllergyRepository)(nil)

type AllergyRepository struct {
	db *gorm.DB
}

func NewAllergyRepository(db *gorm.DB) *AllergyRepository {
	return &AllergyRepository{db: db}
}

func (r *AllergyRepository) FindByID(ctx context.Context, id int64) (*entities.Allergy, error) {
	var m models.AllergyModel

	if err := conn(ctx, r.db).Where("deleted_at IS NULL").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return models.AllergyToEntity(m), nil
}

func (r *UserAllergyRepository) FindByPublicID(
	ctx context.Context,
	publicID string,
) (*entities.UserAllergy, error) {
	var m models.UserAllergyModel

	if err := conn(ctx, r.db).
		Where("public_id = ?", publicID).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return models.UserAllergyToEntity(&m), nil
}

func (r *AllergyRepository) FindAll(ctx context.Context) ([]*entities.Allergy, error) {
	var ms []models.AllergyModel

	if err := conn(ctx, r.db).Find(&ms).Error; err != nil {
		return nil, err
	}

	result := make([]*entities.Allergy, len(ms))
	for i, m := range ms {
		m := m
		result[i] = models.AllergyToEntity(m)
	}

	return result, nil
}

func (r *AllergyRepository) FindByCategory(ctx context.Context, category int) ([]*entities.Allergy, error) {
	var ms []models.AllergyModel

	if err := conn(ctx, r.db).
		Where("category = ?", category).
		Find(&ms).Error; err != nil {
		return nil, err
	}

	result := make([]*entities.Allergy, len(ms))
	for i, m := range ms {
		m := m
		result[i] = models.AllergyToEntity(m)
	}

	return result, nil
}

func (r *AllergyRepository) Save(ctx context.Context, allergy *entities.Allergy) error {
	if allergy == nil {
		return errors.New("allergy is required")
	}

	m := models.AllergyToModel(*allergy)

	if err := conn(ctx, r.db).Create(m).Error; err != nil {
		return err
	}

	allergy.ID = m.ID
	allergy.CreatedAt = m.CreatedAt

	return nil
}

func (r *AllergyRepository) Update(ctx context.Context, allergy *entities.Allergy) error {
	if allergy == nil {
		return errors.New("allergy is required")
	}

	if allergy.ID <= 0 {
		return errors.New("allergy id is required")
	}

	m := models.AllergyToModel(*allergy)

	result := conn(ctx, r.db).
		Model(&models.AllergyModel{}).
		Where("id = ?", allergy.ID).
		Select("*").
		Omit("id", "created_at", "deleted_at").
		Updates(m)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *AllergyRepository) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("allergy id is required")
	}

	result := conn(ctx, r.db).
		Model(&models.AllergyModel{}).
		Where("id = ?", id).
		Updates(map[string]any{"deleted_at": time.Now()})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *AllergyRepository) SaveBatch(
	ctx context.Context,
	allergies []*entities.Allergy,
) error {
	if len(allergies) == 0 {
		return nil
	}

	modelList := make([]models.AllergyModel, 0, len(allergies))

	for _, allergy := range allergies {
		if allergy == nil {
			return errors.New("allergy cannot be nil")
		}

		modelList = append(modelList, *models.AllergyToModel(*allergy))
	}

	if err := conn(ctx, r.db).Create(&modelList).Error; err != nil {
		return err
	}

	for i := range allergies {
		allergies[i].ID = modelList[i].ID
		allergies[i].CreatedAt = modelList[i].CreatedAt
	}

	return nil
}
