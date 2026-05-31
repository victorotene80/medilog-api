package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type OTPCodeRepository struct {
	db *gorm.DB
}

func NewOTPCodeRepository(db *gorm.DB) *OTPCodeRepository {
	return &OTPCodeRepository{db: db}
}

func (r *OTPCodeRepository) FindLatestByRecipientAndPurpose(ctx context.Context, recipient, purpose string) (*entities.OTPCode, error) {
	var m models.OTPCodeModel
	if err := r.db.WithContext(ctx).
		Where("recipient = ? AND purpose = ? AND used_at IS NULL", recipient, purpose).
		Order("created_at DESC").
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.OTPCodeToEntity(&m), nil
}

func (r *OTPCodeRepository) FindByID(ctx context.Context, id int64) (*entities.OTPCode, error) {
	var m models.OTPCodeModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.OTPCodeToEntity(&m), nil
}

func (r *OTPCodeRepository) Save(ctx context.Context, otp *entities.OTPCode) error {
	m := models.OTPCodeToModel(otp)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	otp.ID = m.ID
	return nil
}

func (r *OTPCodeRepository) Update(ctx context.Context, otp *entities.OTPCode) error {
	return r.db.WithContext(ctx).Save(models.OTPCodeToModel(otp)).Error
}

func (r *OTPCodeRepository) InvalidatePreviousByRecipientAndPurpose(ctx context.Context, recipient, purpose string, now time.Time) error {
	return r.db.WithContext(ctx).Model(&models.OTPCodeModel{}).
		Where("recipient = ? AND purpose = ? AND used_at IS NULL", recipient, purpose).
		UpdateColumn("used_at", now).Error
}
