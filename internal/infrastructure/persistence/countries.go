package persistence

import (
	"context"
	"errors"
	"strings"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type CountryRepository struct {
	db *gorm.DB
}

func NewCountryRepository(db *gorm.DB) *CountryRepository {
	return &CountryRepository{db: db}
}

func (r *CountryRepository) FindAll(ctx context.Context) ([]*entities.Country, error) {
	var ms []models.CountryModel

	if err := r.db.WithContext(ctx).
		Order("name ASC").
		Find(&ms).Error; err != nil {
		return nil, err
	}

	result := make([]*entities.Country, len(ms))
	for i, m := range ms {
		m := m
		result[i] = models.CountryToEntity(m)
	}

	return result, nil
}

func (r *CountryRepository) FindByCode(ctx context.Context, code string) (*entities.Country, error) {
	code = strings.TrimSpace(strings.ToUpper(code))

	if code == "" {
		return nil, errors.New("country code is required")
	}

	var m models.CountryModel

	if err := r.db.WithContext(ctx).
		Where("code = ?", code).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return models.CountryToEntity(m), nil
}
