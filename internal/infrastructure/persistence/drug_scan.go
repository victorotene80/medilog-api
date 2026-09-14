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

var _ repository.DrugScanRepository = (*DrugScanRepository)(nil)

type DrugScanRepository struct {
	db *gorm.DB
}

func NewDrugScanRepository(db *gorm.DB) *DrugScanRepository {
	return &DrugScanRepository{db: db}
}

func (r *DrugScanRepository) FindByID(ctx context.Context, id int64) (*entities.DrugScan, error) {
	var m models.DrugScanModel
	if err := conn(ctx, r.db).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.DrugScanToEntity(&m), nil
}

func (r *DrugScanRepository) FindByPublicID(ctx context.Context, userID int64, publicID string) (*entities.DrugScan, error) {
	var m models.DrugScanModel
	err := conn(ctx, r.db).
		Where("user_id = ? AND public_id = ?", userID, publicID).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.DrugScanToEntity(&m), nil
}

func (r *DrugScanRepository) FindByUserID(ctx context.Context, userID int64) ([]*entities.DrugScan, error) {
	var ms []models.DrugScanModel
	if err := conn(ctx, r.db).
		Where("deleted_at IS NULL").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	result := make([]*entities.DrugScan, len(ms))
	for i, m := range ms {
		m := m
		result[i] = models.DrugScanToEntity(&m)
	}
	return result, nil
}

func (r *DrugScanRepository) Save(ctx context.Context, scan *entities.DrugScan) error {
	m := models.DrugScanToModel(scan)
	if err := conn(ctx, r.db).Create(m).Error; err != nil {
		return err
	}
	scan.ID = m.ID
	scan.PublicID = m.PublicID // written back after Postgres generates it
	return nil
}

func (r *DrugScanRepository) Update(ctx context.Context, scan *entities.DrugScan) error {
	if scan == nil {
		return errors.New("drug scan is required")
	}
	if scan.ID <= 0 {
		return errors.New("drug scan id is required")
	}

	model := models.DrugScanToModel(scan)
	result := conn(ctx, r.db).
		Model(&models.DrugScanModel{}).
		Where("id = ?", scan.ID).
		Select("*").
		Omit("id", "public_id", "created_at", "deleted_at").
		Updates(model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *DrugScanRepository) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("drug scan id is required")
	}

	result := conn(ctx, r.db).Model(&models.DrugScanModel{}).Where("id = ?", id).Updates(map[string]any{"deleted_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}
