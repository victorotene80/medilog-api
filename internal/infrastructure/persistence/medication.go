package persistence

import (
	"context"
	"errors"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

var _ repository.MedicationRepository = (*MedicationRepository)(nil)

type MedicationRepository struct {
	db *gorm.DB
	// events drains the aggregate's domain events into the outbox on the same
	// transaction as the write. See drainAggregateEvents.
	events appContracts.MessagePublisher
}

func NewMedicationRepository(db *gorm.DB, events appContracts.MessagePublisher) *MedicationRepository {
	return &MedicationRepository{db: db, events: events}
}

func (r *MedicationRepository) FindByID(ctx context.Context, id int64) (*aggregates.MedicationAggregate, error) {
	var m models.MedicationModel
	if err := conn(ctx, r.db).Where("deleted_at IS NULL").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.hydrate(ctx, &m)
}

func (r *MedicationRepository) FindByUserID(ctx context.Context, userID int64) ([]*aggregates.MedicationAggregate, error) {
	var ms []models.MedicationModel
	if err := conn(ctx, r.db).Where("deleted_at IS NULL").Where("user_id = ?", userID).Find(&ms).Error; err != nil {
		return nil, err
	}
	return r.hydrateAll(ctx, ms)
}

func (r *MedicationRepository) FindActiveByUserID(ctx context.Context, userID int64, now time.Time) ([]*aggregates.MedicationAggregate, error) {
	var ms []models.MedicationModel
	if err := conn(ctx, r.db).
		Where("deleted_at IS NULL").
		Where(
			"user_id = ? AND is_completed = false"+
				" AND (start_date IS NULL OR start_date <= ?)"+
				" AND (end_date IS NULL OR end_date > ?)",
			userID, now, now,
		).
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return r.hydrateAll(ctx, ms)
}

func (r *MedicationRepository) Save(ctx context.Context, agg *aggregates.MedicationAggregate) error {
	return conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
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
		// Drained here, on the same transaction as the write: leaving it to the
		// caller meant most mutations raised events that were silently discarded.
		return drainAggregateEvents(ctx, tx, r.events, agg)
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
	// Wrapped in a transaction so the outbox envelope commits with the row it
	// describes; a bare Updates cannot carry the drain.
	return conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		result := tx.
			Model(&models.MedicationModel{}).
			Where("id = ?", agg.Medication.ID).
			Select("*").
			Omit("id", "public_id", "created_at", "deleted_at").
			Updates(model)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return repository.ErrNotFound
		}

		return drainAggregateEvents(ctx, tx, r.events, agg)
	})
}

func (r *MedicationRepository) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("medication id is required")
	}

	return conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.MedicationAdherenceLogModel{}).
			Where("medication_id = ?", id).
			Updates(map[string]any{"deleted_at": time.Now()}).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.MedicationTimeModel{}).
			Where("medication_id = ?", id).
			Updates(map[string]any{"deleted_at": time.Now()}).Error; err != nil {
			return err
		}

		result := tx.Model(&models.MedicationModel{}).Where("id = ?", id).Updates(map[string]any{"deleted_at": time.Now()})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return repository.ErrNotFound
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
	// Checked: dropping this error rendered a medication with no dose times as a
	// complete 200, which on a reminder product reads as "no doses scheduled".
	if err := conn(ctx, r.db).Where("medication_id = ?", m.ID).Find(&timeModels).Error; err != nil {
		return nil, err
	}
	times := make([]*entities.MedicationTime, len(timeModels))
	for i, t := range timeModels {
		t := t
		times[i] = models.MedicationTimeToEntity(&t)
	}

	var logModels []models.MedicationAdherenceLogModel
	if err := conn(ctx, r.db).Where("medication_id = ?", m.ID).Find(&logModels).Error; err != nil {
		return nil, err
	}
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
	if err := conn(ctx, r.db).
		Where("user_id = ? AND public_id = ?", userID, publicID).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.hydrate(ctx, &m)
}
