package persistence

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type MedicationTimeRepository struct {
	db *gorm.DB
}

func NewMedicationTimeRepository(db *gorm.DB) *MedicationTimeRepository {
	return &MedicationTimeRepository{db: db}
}

func (r *MedicationTimeRepository) FindByMedicationID(ctx context.Context, medicationID int64) ([]*entities.MedicationTime, error) {
	var ms []models.MedicationTimeModel
	if err := r.db.WithContext(ctx).Where("medication_id = ?", medicationID).Find(&ms).Error; err != nil {
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
	if err := r.db.WithContext(ctx).Create(&ms).Error; err != nil {
		return err
	}
	for i, m := range ms {
		times[i].ID = m.ID
	}
	return nil
}

func (r *MedicationTimeRepository) DeleteByMedicationID(ctx context.Context, medicationID int64) error {
	return r.db.WithContext(ctx).Where("medication_id = ?", medicationID).Delete(&models.MedicationTimeModel{}).Error
}
