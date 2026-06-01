package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type MedicationRepository struct {
	db *gorm.DB
}

func NewMedicationRepository(db *gorm.DB) *MedicationRepository {
	return &MedicationRepository{db: db}
}

func (r *MedicationRepository) FindByID(ctx context.Context, id int64) (*aggregates.MedicationAggregate, error) {
	var m models.MedicationModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.hydrate(ctx, &m)
}

func (r *MedicationRepository) FindByUserID(ctx context.Context, userID int64) ([]*aggregates.MedicationAggregate, error) {
	var ms []models.MedicationModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&ms).Error; err != nil {
		return nil, err
	}
	return r.hydrateAll(ctx, ms)
}

func (r *MedicationRepository) FindActiveByUserID(ctx context.Context, userID int64, now time.Time) ([]*aggregates.MedicationAggregate, error) {
	var ms []models.MedicationModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_completed = false AND (end_date IS NULL OR end_date > ?)", userID, now).
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return r.hydrateAll(ctx, ms)
}

func (r *MedicationRepository) Save(ctx context.Context, agg *aggregates.MedicationAggregate) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		medModel := models.MedicationToModel(agg.Medication)
		if err := tx.Create(medModel).Error; err != nil {
			return err
		}
		agg.Medication.ID = medModel.ID

		for _, t := range agg.Times {
			t.MedicationID = medModel.ID
			timeModel := models.MedicationTimeToModel(t)
			if err := tx.Create(timeModel).Error; err != nil {
				return err
			}
			t.ID = timeModel.ID
		}
		return nil
	})
}

func (r *MedicationRepository) Update(ctx context.Context, agg *aggregates.MedicationAggregate) error {
	if agg == nil || agg.Medication == nil {
		return errors.New("medication is required")
	}
	if agg.Medication.ID <= 0 {
		return errors.New("medication id is required")
	}

	model := models.MedicationToModel(agg.Medication)
	result := r.db.WithContext(ctx).
		Model(&models.MedicationModel{}).
		Where("id = ?", agg.Medication.ID).
		Select("*").
		Omit("id", "public_id", "created_at", "deleted_at").
		Updates(model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *MedicationRepository) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("medication id is required")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("medication_id = ?", id).
			Delete(&models.MedicationAdherenceLogModel{}).Error; err != nil {
			return err
		}

		if err := tx.Where("medication_id = ?", id).
			Delete(&models.MedicationTimeModel{}).Error; err != nil {
			return err
		}

		result := tx.Delete(&models.MedicationModel{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *MedicationRepository) hydrate(ctx context.Context, m *models.MedicationModel) (*aggregates.MedicationAggregate, error) {
	entity, err := models.MedicationToEntity(m)
	if err != nil {
		return nil, err
	}

	var timeModels []models.MedicationTimeModel
	r.db.WithContext(ctx).Where("medication_id = ?", m.ID).Find(&timeModels)
	times := make([]*entities.MedicationTime, len(timeModels))
	for i, t := range timeModels {
		t := t
		times[i] = models.MedicationTimeToEntity(&t)
	}

	var logModels []models.MedicationAdherenceLogModel
	r.db.WithContext(ctx).Where("medication_id = ?", m.ID).Find(&logModels)
	logs := make([]*entities.MedicationAdherenceLog, 0, len(logModels))
	for _, l := range logModels {
		l := l
		log, err := models.MedicationAdherenceLogToEntity(&l)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return aggregates.RestoreMedicationAggregate(entity, times, logs, 0), nil
}

func (r *MedicationRepository) hydrateAll(ctx context.Context, ms []models.MedicationModel) ([]*aggregates.MedicationAggregate, error) {
	result := make([]*aggregates.MedicationAggregate, 0, len(ms))
	for _, m := range ms {
		m := m
		agg, err := r.hydrate(ctx, &m)
		if err != nil {
			return nil, err
		}
		result = append(result, agg)
	}
	return result, nil
}

func (r *MedicationRepository) FindByPublicID(ctx context.Context, userID int64, publicID string) (*aggregates.MedicationAggregate, error) {
	var m models.MedicationModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND public_id = ?", userID, publicID).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.hydrate(ctx, &m)
}
