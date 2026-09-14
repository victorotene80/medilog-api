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

var _ repository.MedicationTimeRepository = (*MedicationTimeRepository)(nil)

type MedicationTimeRepository struct {
	db *gorm.DB
}

func NewMedicationTimeRepository(db *gorm.DB) *MedicationTimeRepository {
	return &MedicationTimeRepository{db: db}
}

func (r *MedicationTimeRepository) FindByMedicationID(ctx context.Context, medicationID int64) ([]*entities.MedicationTime, error) {
	var ms []models.MedicationTimeModel
	if err := conn(ctx, r.db).Where("medication_id = ?", medicationID).Find(&ms).Error; err != nil {
		return nil, err
	}
	result := make([]*entities.MedicationTime, len(ms))
	for i, m := range ms {
		m := m
		result[i] = models.MedicationTimeToEntity(&m)
	}
	return result, nil
}

func (r *MedicationTimeRepository) SaveAll(ctx context.Context, times []*entities.MedicationTime) error {
	ms := make([]models.MedicationTimeModel, len(times))
	for i, t := range times {
		ms[i] = *models.MedicationTimeToModel(t)
	}
	if err := conn(ctx, r.db).Create(&ms).Error; err != nil {
		return err
	}
	for i, m := range ms {
		times[i].ID = m.ID
	}
	return nil
}

func (r *MedicationTimeRepository) DeleteByMedicationID(ctx context.Context, medicationID int64) error {
	if medicationID <= 0 {
		return errors.New("medication id is required")
	}
	return conn(ctx, r.db).Model(&models.MedicationTimeModel{}).Where("medication_id = ?", medicationID).Updates(map[string]any{"deleted_at": time.Now()}).Error
}

func (r *MedicationTimeRepository) ReplaceTimes(ctx context.Context, medicationID int64, times []*entities.MedicationTime) error {
	if medicationID <= 0 {
		return errors.New("medication id is required")
	}
	return conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.MedicationTimeModel{}).
			Where("medication_id = ?", medicationID).
			Updates(map[string]any{"deleted_at": time.Now()}).Error; err != nil {
			return err
		}
		if len(times) > 0 {
			ms := make([]models.MedicationTimeModel, len(times))
			for i, t := range times {
				ms[i] = *models.MedicationTimeToModel(t)
			}
			if err := tx.Create(&ms).Error; err != nil {
				return err
			}
			for i, m := range ms {
				times[i].ID = m.ID
			}
		}
		return nil
	})
}
