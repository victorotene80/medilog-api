package persistence

import (
	"context"
	"errors"
	"strings"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type UserAllergyRepository struct {
	db *gorm.DB
}

func NewUserAllergyRepository(db *gorm.DB) *UserAllergyRepository {
	return &UserAllergyRepository{db: db}
}

func (r *UserAllergyRepository) FindByID(ctx context.Context, id int64) (*entities.UserAllergy, error) {
	var m models.UserAllergyModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.UserAllergyToEntity(&m), nil
}

func (r *UserAllergyRepository) FindByUserID(ctx context.Context, userID int64) ([]*entities.UserAllergy, error) {
	var ms []models.UserAllergyModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&ms).Error; err != nil {
		return nil, err
	}
	result := make([]*entities.UserAllergy, len(ms))
	for i, m := range ms {
		m := m
		result[i] = models.UserAllergyToEntity(&m)
	}
	return result, nil
}

func (r *UserAllergyRepository) Save(ctx context.Context, allergy *entities.UserAllergy) error {
	m := models.UserAllergyEntityToModel(allergy)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	allergy.ID = m.ID
	allergy.PublicID = m.PublicID
	allergy.CreatedAt = m.CreatedAt
	return nil
}

func (r *UserAllergyRepository) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("user allergy id is required")
	}

	result := r.db.WithContext(ctx).Delete(&models.UserAllergyModel{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserAllergyRepository) SaveBatch(
	ctx context.Context,
	allergies []*entities.UserAllergy,
) error {
	if len(allergies) == 0 {
		return nil
	}

	modelList := make([]models.UserAllergyModel, 0, len(allergies))

	for _, allergy := range allergies {
		if allergy == nil {
			return errors.New("user allergy cannot be nil")
		}

		model := models.UserAllergyEntityToModel(allergy)
		modelList = append(modelList, *model)
	}

	if err := r.db.WithContext(ctx).Create(&modelList).Error; err != nil {
		return err
	}

	for i := range allergies {
		allergies[i].ID = modelList[i].ID
		allergies[i].PublicID = modelList[i].PublicID
		allergies[i].CreatedAt = modelList[i].CreatedAt
	}

	return nil
}

func (r *UserAllergyRepository) DeleteByPublicID(
	ctx context.Context,
	userID int64,
	publicID string,
) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND public_id = ?", userID, publicID).
		Delete(&models.UserAllergyModel{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *UserAllergyRepository) FindByUserIDAndAllergyID(
	ctx context.Context,
	userID int64,
	allergyID int64,
) (*entities.UserAllergy, error) {
	if userID <= 0 {
		return nil, errors.New("user id is required")
	}

	if allergyID <= 0 {
		return nil, errors.New("allergy id is required")
	}

	var m models.UserAllergyModel

	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND allergy_id = ?", userID, allergyID).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return models.UserAllergyToEntity(&m), nil
}

func (r *UserAllergyRepository) FindByUserIDAndName(
	ctx context.Context,
	userID int64,
	name string,
) (*entities.UserAllergy, error) {
	if userID <= 0 {
		return nil, errors.New("user id is required")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("allergy name is required")
	}

	var m models.UserAllergyModel

	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND LOWER(name) = LOWER(?)", userID, name).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return models.UserAllergyToEntity(&m), nil
}
