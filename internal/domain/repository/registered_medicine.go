package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type RegisteredMedicineRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.RegisteredMedicine, error)
	FindByRegistrationNumber(ctx context.Context, registrationNumber, countryCode string) (*entities.RegisteredMedicine, error)
	FindByBarcode(ctx context.Context, barcode string) (*entities.RegisteredMedicine, error)
	Search(ctx context.Context, drugName, countryCode string, limit int) ([]*entities.RegisteredMedicine, error)
	Save(ctx context.Context, m *entities.RegisteredMedicine) error
	Update(ctx context.Context, m *entities.RegisteredMedicine) error
	Delete(ctx context.Context, id int64) error
}
