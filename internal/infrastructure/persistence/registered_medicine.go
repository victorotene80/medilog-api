package persistence

import (
	"context"
	"errors"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type RegisteredMedicineRepository struct {
	db *gorm.DB
}

func NewRegisteredMedicineRepository(db *gorm.DB) *RegisteredMedicineRepository {
	return &RegisteredMedicineRepository{db: db}
}

func (r *RegisteredMedicineRepository) FindByID(ctx context.Context, id int64) (*entities.RegisteredMedicine, error) {
	var m models.RegisteredMedicineModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.RegisteredMedicineToEntity(&m), nil
}

func (r *RegisteredMedicineRepository) FindByRegistrationNumber(
	ctx context.Context,
	registrationNumber, countryCode string,
) (*entities.RegisteredMedicine, error) {
	var m models.RegisteredMedicineModel
	err := r.db.WithContext(ctx).
		Where("registration_number = ? AND country_code = ?", registrationNumber, countryCode).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.RegisteredMedicineToEntity(&m), nil
}

func (r *RegisteredMedicineRepository) FindByBarcode(ctx context.Context, barcode string) (*entities.RegisteredMedicine, error) {
	var m models.RegisteredMedicineModel
	if err := r.db.WithContext(ctx).Where("barcode = ?", barcode).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.RegisteredMedicineToEntity(&m), nil
}

func (r *RegisteredMedicineRepository) Search(
	ctx context.Context,
	drugName, countryCode string,
	limit int,
) ([]*entities.RegisteredMedicine, error) {
	var ms []models.RegisteredMedicineModel
	q := r.db.WithContext(ctx).
		Where("country_code = ? AND drug_name ILIKE ?", countryCode, "%"+drugName+"%").
		Limit(limit).
		Find(&ms)
	if q.Error != nil {
		return nil, q.Error
	}
	result := make([]*entities.RegisteredMedicine, len(ms))
	for i, m := range ms {
		m := m
		result[i] = models.RegisteredMedicineToEntity(&m)
	}
	return result, nil
}

func (r *RegisteredMedicineRepository) Save(ctx context.Context, e *entities.RegisteredMedicine) error {
	m := models.RegisteredMedicineToModel(e)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	e.ID = m.ID
	return nil
}

func (r *RegisteredMedicineRepository) Update(ctx context.Context, e *entities.RegisteredMedicine) error {
	return r.db.WithContext(ctx).Save(models.RegisteredMedicineToModel(e)).Error
}

func (r *RegisteredMedicineRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.RegisteredMedicineModel{}, id).Error
}
