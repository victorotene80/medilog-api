package repository

import (
	"context"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type VisitRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.Visit, error)
	FindByPublicID(ctx context.Context, userID int64, publicID string) (*entities.Visit, error)
	FindByUserID(ctx context.Context, userID int64) ([]*entities.Visit, error)
	FindByUserIDAndDateRange(ctx context.Context, userID int64, from, to time.Time) ([]*entities.Visit, error)
	Save(ctx context.Context, visit *entities.Visit) error
	Update(ctx context.Context, visit *entities.Visit) error
	Delete(ctx context.Context, id int64) error
}
