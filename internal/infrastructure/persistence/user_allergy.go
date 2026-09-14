package persistence

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

var _ repository.UserAllergyRepository = (*UserAllergyRepository)(nil)

type UserAllergyRepository struct {
	db *gorm.DB
}

func NewUserAllergyRepository(db *gorm.DB) *UserAllergyRepository {
	return &UserAllergyRepository{db: db}
}

func (r *UserAllergyRepository) FindByID(ctx context.Context, id int64) (*entities.UserAllergy, error) {
	var m models.UserAllergyModel
	if err := conn(ctx, r.db).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.UserAllergyToEntity(&m), nil
}

func (r *UserAllergyRepository) FindByUserID(ctx context.Context, userID int64) ([]*entities.UserAllergy, error) {
	var ms []models.UserAllergyModel
	if err := conn(ctx, r.db).Where("deleted_at IS NULL").Where("user_id = ?", userID).Find(&ms).Error; err != nil {
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
	if err := conn(ctx, r.db).Create(m).Error; err != nil {
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

	result := conn(ctx, r.db).Model(&models.UserAllergyModel{}).Where("id = ?", id).Updates(map[string]any{"deleted_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
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

	if err := conn(ctx, r.db).Create(&modelList).Error; err != nil {
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
	result := conn(ctx, r.db).
		Model(&models.UserAllergyModel{}).
		Where("user_id = ? AND public_id = ?", userID, publicID).
		Updates(map[string]any{"deleted_at": time.Now()})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
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

	if err := conn(ctx, r.db).
		Where("user_id = ? AND allergy_id = ?", userID, allergyID).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return models.UserAllergyToEntity(&m), nil
}

// Update rewrites the mutable columns of an existing user allergy. public_id
// and created_at are omitted so an update can never re-key or re-date a row.
func (r *UserAllergyRepository) Update(
	ctx context.Context,
	allergy *entities.UserAllergy,
) error {
	if allergy == nil {
		return errors.New("user allergy is required")
	}

	if allergy.ID <= 0 {
		return errors.New("user allergy id is required")
	}

	m := models.UserAllergyEntityToModel(allergy)

	result := conn(ctx, r.db).
		Model(&models.UserAllergyModel{}).
		Where("id = ?", allergy.ID).
		Select("*").
		Omit("id", "public_id", "created_at", "deleted_at").
		Updates(m)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
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

	if err := conn(ctx, r.db).
		Where("user_id = ? AND LOWER(name) = LOWER(?)", userID, name).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return models.UserAllergyToEntity(&m), nil
}
