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

var _ repository.MedicationAdherenceLogRepository = (*MedicationAdherenceLogRepository)(nil)

type MedicationAdherenceLogRepository struct {
	db *gorm.DB
}

func NewMedicationAdherenceLogRepository(db *gorm.DB) *MedicationAdherenceLogRepository {
	return &MedicationAdherenceLogRepository{db: db}
}

func (r *MedicationAdherenceLogRepository) FindByID(ctx context.Context, id int64) (*entities.MedicationAdherenceLog, error) {
	var m models.MedicationAdherenceLogModel
	if err := conn(ctx, r.db).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.MedicationAdherenceLogToEntity(&m)
}

func (r *MedicationAdherenceLogRepository) FindByMedicationID(ctx context.Context, medicationID int64) ([]*entities.MedicationAdherenceLog, error) {
	var ms []models.MedicationAdherenceLogModel
	if err := conn(ctx, r.db).Where("medication_id = ?", medicationID).Find(&ms).Error; err != nil {
		return nil, err
	}
	return toAdherenceEntities(ms)
}

func (r *MedicationAdherenceLogRepository) FindByUserIDAndDateRange(ctx context.Context, userID int64, from, to time.Time) ([]*entities.MedicationAdherenceLog, error) {
	var ms []models.MedicationAdherenceLogModel
	if err := conn(ctx, r.db).
		Where("user_id = ? AND scheduled_at BETWEEN ? AND ?", userID, from, to).
		Order("scheduled_at ASC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return toAdherenceEntities(ms)
}

func (r *MedicationAdherenceLogRepository) Save(ctx context.Context, log *entities.MedicationAdherenceLog) error {
	m := models.MedicationAdherenceLogToModel(log)
	if err := conn(ctx, r.db).Create(m).Error; err != nil {
		return err
	}
	log.ID = m.ID
	return nil
}

func (r *MedicationAdherenceLogRepository) Update(ctx context.Context, log *entities.MedicationAdherenceLog) error {
	if log == nil {
		return errors.New("medication adherence log is required")
	}
	if log.ID <= 0 {
		return errors.New("medication adherence log id is required")
	}

	model := models.MedicationAdherenceLogToModel(log)
	result := conn(ctx, r.db).
		Model(&models.MedicationAdherenceLogModel{}).
		Where("id = ?", log.ID).
		Select("*").
		Omit("id", "created_at", "deleted_at").
		Updates(model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
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
