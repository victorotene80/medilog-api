package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type MedicationAdherenceLogRepository struct {
	db *gorm.DB
}

func NewMedicationAdherenceLogRepository(db *gorm.DB) *MedicationAdherenceLogRepository {
	return &MedicationAdherenceLogRepository{db: db}
}

func (r *MedicationAdherenceLogRepository) FindByID(ctx context.Context, id int64) (*entities.MedicationAdherenceLog, error) {
	var m models.MedicationAdherenceLogModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.MedicationAdherenceLogToEntity(&m)
}

func (r *MedicationAdherenceLogRepository) FindByMedicationID(ctx context.Context, medicationID int64) ([]*entities.MedicationAdherenceLog, error) {
	var ms []models.MedicationAdherenceLogModel
	if err := r.db.WithContext(ctx).Where("medication_id = ?", medicationID).Find(&ms).Error; err != nil {
		return nil, err
	}
	return toAdherenceEntities(ms)
}

func (r *MedicationAdherenceLogRepository) FindByUserIDAndDateRange(ctx context.Context, userID int64, from, to time.Time) ([]*entities.MedicationAdherenceLog, error) {
	var ms []models.MedicationAdherenceLogModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND scheduled_at BETWEEN ? AND ?", userID, from, to).
		Order("scheduled_at ASC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return toAdherenceEntities(ms)
}

func (r *MedicationAdherenceLogRepository) Save(ctx context.Context, log *entities.MedicationAdherenceLog) error {
	m := models.MedicationAdherenceLogToModel(log)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	log.ID = m.ID
	return nil
}

func (r *MedicationAdherenceLogRepository) Update(ctx context.Context, log *entities.MedicationAdherenceLog) error {
	return r.db.WithContext(ctx).Save(models.MedicationAdherenceLogToModel(log)).Error
}

func toAdherenceEntities(ms []models.MedicationAdherenceLogModel) ([]*entities.MedicationAdherenceLog, error) {
	result := make([]*entities.MedicationAdherenceLog, 0, len(ms))
	for _, m := range ms {
		m := m
		e, err := models.MedicationAdherenceLogToEntity(&m)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}
