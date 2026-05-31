package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type AllergyRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.Allergy, error)
	FindAll(ctx context.Context) ([]*entities.Allergy, error)
	FindByCategory(ctx context.Context, category int) ([]*entities.Allergy, error)
	SaveBatch(ctx context.Context, allergies []*entities.Allergy) error
	Save(ctx context.Context, allergy *entities.Allergy) error
	Update(ctx context.Context, allergy *entities.Allergy) error
	Delete(ctx context.Context, id int64) error
}
