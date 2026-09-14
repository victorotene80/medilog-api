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

var _ repository.VisitRepository = (*VisitRepository)(nil)

type VisitRepository struct {
	db *gorm.DB
}

func NewVisitRepository(db *gorm.DB) *VisitRepository {
	return &VisitRepository{db: db}
}

func (r *VisitRepository) FindByID(ctx context.Context, id int64) (*entities.Visit, error) {
	var m models.VisitModel
	if err := conn(ctx, r.db).Where("deleted_at IS NULL").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.VisitToEntity(&m), nil
}

func (r *VisitRepository) FindByPublicID(ctx context.Context, userID int64, publicID string) (*entities.Visit, error) {
	var m models.VisitModel
	if err := conn(ctx, r.db).
		Where("user_id = ? AND public_id = ?", userID, publicID).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.VisitToEntity(&m), nil
}

func (r *VisitRepository) FindByUserID(ctx context.Context, userID int64) ([]*entities.Visit, error) {
	var ms []models.VisitModel
	if err := conn(ctx, r.db).Where("deleted_at IS NULL").Where("user_id = ?", userID).Order("visit_date DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	result := make([]*entities.Visit, len(ms))
	for i, m := range ms {
		m := m
		result[i] = models.VisitToEntity(&m)
	}
	return result, nil
}

func (r *VisitRepository) FindByUserIDAndDateRange(ctx context.Context, userID int64, from, to time.Time) ([]*entities.Visit, error) {
	var ms []models.VisitModel
	if err := conn(ctx, r.db).
		Where("deleted_at IS NULL").
		Where("user_id = ? AND visit_date BETWEEN ? AND ?", userID, from, to).
		Order("visit_date DESC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	result := make([]*entities.Visit, len(ms))
	for i, m := range ms {
		m := m
		result[i] = models.VisitToEntity(&m)
	}
	return result, nil
}

func (r *VisitRepository) Save(ctx context.Context, visit *entities.Visit) error {
	m := models.VisitToModel(visit)
	if err := conn(ctx, r.db).Create(m).Error; err != nil {
		return err
	}
	visit.ID = m.ID
	visit.PublicID = m.PublicID
	return nil
}

func (r *VisitRepository) Update(ctx context.Context, visit *entities.Visit) error {
	if visit.ID == 0 {
		return errors.New("visit id is required")
	}
	if visit.UserID == 0 {
		return errors.New("visit user id is required")
	}

	result := conn(ctx, r.db).
		Model(&models.VisitModel{}).
		Where("id = ? AND user_id = ?", visit.ID, visit.UserID).
		Updates(map[string]any{
			"hospital_name":   visit.HospitalName,
			"diagnosis":       visit.Diagnosis,
			"visit_date":      visit.VisitDate,
			"outcome":         visit.Outcome,
			"meds_count":      visit.MedsCount,
			"doctor":          visit.Doctor,
			"chief_complaint": visit.ChiefComplaint,
			"notes":           visit.Notes,
			"blood_pressure":  visit.BloodPressure,
			"temperature":     visit.Temperature,
			"weight":          visit.Weight,
			"pulse":           visit.Pulse,
			"updated_at":      visit.UpdatedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *VisitRepository) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("visit id is required")
	}

	result := conn(ctx, r.db).Model(&models.VisitModel{}).Where("id = ?", id).Updates(map[string]any{"deleted_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}
