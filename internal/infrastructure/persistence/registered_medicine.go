package persistence

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

var _ repository.RegisteredMedicineRepository = (*RegisteredMedicineRepository)(nil)

type RegisteredMedicineRepository struct {
	db *gorm.DB
}

func NewRegisteredMedicineRepository(db *gorm.DB) *RegisteredMedicineRepository {
	return &RegisteredMedicineRepository{db: db}
}

func (r *RegisteredMedicineRepository) FindByID(ctx context.Context, id int64) (*entities.RegisteredMedicine, error) {
	var m models.RegisteredMedicineModel
	if err := conn(ctx, r.db).First(&m, id).Error; err != nil {
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
	err := conn(ctx, r.db).
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
	if err := conn(ctx, r.db).Where("barcode = ?", barcode).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.RegisteredMedicineToEntity(&m), nil
}

// likeEscaper keeps user-typed "%" and "_" literal; unescaped, a bare "%"
// matches every row in the country.
var likeEscaper = strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`)

func (r *RegisteredMedicineRepository) Search(
	ctx context.Context,
	drugName, countryCode string,
	limit int,
) ([]*entities.RegisteredMedicine, error) {
	var ms []models.RegisteredMedicineModel
	q := conn(ctx, r.db).
		Where(`country_code = ? AND drug_name ILIKE ? ESCAPE '\'`, countryCode, "%"+likeEscaper.Replace(drugName)+"%").
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
	if err := conn(ctx, r.db).Create(m).Error; err != nil {
		return err
	}
	e.ID = m.ID
	return nil
}

func (r *RegisteredMedicineRepository) Update(ctx context.Context, e *entities.RegisteredMedicine) error {
	if e == nil {
		return errors.New("registered medicine is required")
	}
	if e.ID <= 0 {
		return errors.New("registered medicine id is required")
	}

	model := models.RegisteredMedicineToModel(e)
	result := conn(ctx, r.db).
		Model(&models.RegisteredMedicineModel{}).
		Where("id = ?", e.ID).
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

func (r *RegisteredMedicineRepository) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("registered medicine id is required")
	}

	result := conn(ctx, r.db).Model(&models.RegisteredMedicineModel{}).Where("id = ?", id).Updates(map[string]any{"deleted_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}
