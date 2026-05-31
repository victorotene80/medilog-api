package persistence

import (
	"context"
	"errors"
	"strings"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type EmergencyContactRepository struct {
	db *gorm.DB
}

func NewEmergencyContactRepository(db *gorm.DB) *EmergencyContactRepository {
	return &EmergencyContactRepository{db: db}
}

func (r *EmergencyContactRepository) FindByID(
	ctx context.Context,
	id int64,
) (*entities.EmergencyContact, error) {
	var m models.EmergencyContactModel

	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return models.EmergencyContactToEntity(&m), nil
}

func (r *EmergencyContactRepository) FindByPublicID(
	ctx context.Context,
	publicID string,
) (*entities.EmergencyContact, error) {
	publicID = strings.TrimSpace(publicID)
	if publicID == "" {
		return nil, errors.New("public id is required")
	}

	var m models.EmergencyContactModel

	if err := r.db.WithContext(ctx).
		Where("public_id = ?", publicID).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return models.EmergencyContactToEntity(&m), nil
}

func (r *EmergencyContactRepository) FindByUserID(
	ctx context.Context,
	userID int64,
) ([]*entities.EmergencyContact, error) {
	var ms []models.EmergencyContactModel

	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id ASC").
		Find(&ms).Error; err != nil {
		return nil, err
	}

	result := make([]*entities.EmergencyContact, len(ms))
	for i, m := range ms {
		m := m
		result[i] = models.EmergencyContactToEntity(&m)
	}

	return result, nil
}

func (r *EmergencyContactRepository) Save(
	ctx context.Context,
	contact *entities.EmergencyContact,
) error {
	if contact == nil {
		return errors.New("emergency contact is required")
	}

	m := models.EmergencyContactEntityToModel(contact)

	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}

	contact.ID = m.ID
	contact.PublicID = m.PublicID
	contact.CreatedAt = m.CreatedAt
	contact.UpdatedAt = m.UpdatedAt

	return nil
}

func (r *EmergencyContactRepository) Update(
	ctx context.Context,
	contact *entities.EmergencyContact,
) error {
	if contact == nil {
		return errors.New("emergency contact is required")
	}

	if contact.ID <= 0 {
		return errors.New("emergency contact id is required")
	}

	m := models.EmergencyContactEntityToModel(contact)

	result := r.db.WithContext(ctx).
		Model(&models.EmergencyContactModel{}).
		Where("id = ?", contact.ID).
		Select("*").
		Omit("id", "public_id", "created_at").
		Updates(m)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *EmergencyContactRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return errors.New("emergency contact id is required")
	}

	result := r.db.WithContext(ctx).
		Delete(&models.EmergencyContactModel{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *EmergencyContactRepository) DeleteByPublicID(
	ctx context.Context,
	userID int64,
	publicID string,
) error {
	publicID = strings.TrimSpace(publicID)
	if publicID == "" {
		return errors.New("public id is required")
	}

	if userID <= 0 {
		return errors.New("user id is required")
	}

	result := r.db.WithContext(ctx).
		Where("user_id = ? AND public_id = ?", userID, publicID).
		Delete(&models.EmergencyContactModel{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *EmergencyContactRepository) FindByUserIDAndPhone(
	ctx context.Context,
	userID int64,
	phone string,
) (*entities.EmergencyContact, error) {
	if userID <= 0 {
		return nil, errors.New("user id is required")
	}

	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, errors.New("phone is required")
	}

	var m models.EmergencyContactModel

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND phone = ?", userID, phone).
		First(&m).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return models.EmergencyContactToEntity(&m), nil
}
