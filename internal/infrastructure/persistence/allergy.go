package persistence

import (
	"context"
	"errors"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type AllergyRepository struct {
	db *gorm.DB
}

func NewAllergyRepository(db *gorm.DB) *AllergyRepository {
	return &AllergyRepository{db: db}
}

func (r *AllergyRepository) FindByID(ctx context.Context, id int64) (*entities.Allergy, error) {
	var m models.AllergyModel

	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
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

	if err := r.db.WithContext(ctx).
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

	if err := r.db.WithContext(ctx).Find(&ms).Error; err != nil {
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

	if err := r.db.WithContext(ctx).
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

	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
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

	result := r.db.WithContext(ctx).
		Model(&models.AllergyModel{}).
		Where("id = ?", allergy.ID).
		Select("*").
		Omit("id", "created_at").
		Updates(m)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *AllergyRepository) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("allergy id is required")
	}

	result := r.db.WithContext(ctx).
		Delete(&models.AllergyModel{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
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

	if err := r.db.WithContext(ctx).Create(&modelList).Error; err != nil {
		return err
	}

	for i := range allergies {
		allergies[i].ID = modelList[i].ID
		allergies[i].CreatedAt = modelList[i].CreatedAt
	}

	return nil
}
