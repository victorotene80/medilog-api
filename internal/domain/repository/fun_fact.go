package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type FunFactRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.FunFact, error)
	FindAll(ctx context.Context, activeOnly bool) ([]*entities.FunFact, error)
	Save(ctx context.Context, fact *entities.FunFact) error
	Update(ctx context.Context, fact *entities.FunFact) error
	Delete(ctx context.Context, id int64) error
}
