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

var _ repository.EmergencyContactRepository = (*EmergencyContactRepository)(nil)

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

	if err := conn(ctx, r.db).First(&m, id).Error; err != nil {
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

	if err := conn(ctx, r.db).
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

	if err := conn(ctx, r.db).
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

	if err := conn(ctx, r.db).Create(m).Error; err != nil {
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

	result := conn(ctx, r.db).
		Model(&models.EmergencyContactModel{}).
		Where("id = ?", contact.ID).
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

// SetPrimary promotes contactID to be the user's only primary contact.
//
// The demote and promote are deliberately two statements inside one
// transaction. A single `SET is_primary = (id = ?)` would be shorter but
// Postgres checks a non-deferrable unique index once per updated row, so
// depending on the order rows happen to be processed the statement can
// transiently hold two primaries and trip the one-primary-per-user index.
// A unique index cannot be declared DEFERRABLE, so ordering is the only fix.
func (r *EmergencyContactRepository) SetPrimary(
	ctx context.Context,
	userID, contactID int64,
	now time.Time,
) error {
	if userID <= 0 {
		return errors.New("user id is required")
	}

	if contactID <= 0 {
		return errors.New("emergency contact id is required")
	}

	return conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.EmergencyContactModel{}).
			Where("user_id = ? AND is_primary = ? AND id <> ?", userID, true, contactID).
			Updates(map[string]any{"is_primary": false, "updated_at": now}).
			Error; err != nil {
			return err
		}

		result := tx.Model(&models.EmergencyContactModel{}).
			Where("user_id = ? AND id = ?", userID, contactID).
			Updates(map[string]any{"is_primary": true, "updated_at": now})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return repository.ErrNotFound
		}

		return nil
	})
}

func (r *EmergencyContactRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return errors.New("emergency contact id is required")
	}

	// Explicit soft delete rather than db.Delete: gorm's Delete is only a soft
	// delete because EmergencyContactModel.DeletedAt happens to be
	// gorm.DeletedAt. models/user.go declares the same field as *time.Time,
	// for which the identical call is a hard DELETE — and nothing would fail at
	// compile time if this model were ever normalised the same way.
	result := conn(ctx, r.db).
		Model(&models.EmergencyContactModel{}).
		Where("id = ?", id).
		Updates(map[string]any{"deleted_at": time.Now().UTC()})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
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

	result := conn(ctx, r.db).
		Model(&models.EmergencyContactModel{}).
		Where("user_id = ? AND public_id = ?", userID, publicID).
		Updates(map[string]any{"deleted_at": time.Now().UTC()})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
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

	err := conn(ctx, r.db).
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
