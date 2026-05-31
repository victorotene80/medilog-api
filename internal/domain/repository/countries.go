package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type CountryRepository interface {
	FindAll(ctx context.Context) ([]*entities.Country, error)
	FindByCode(ctx context.Context, code string) (*entities.Country, error)
}
