package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type DrugScanRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.DrugScan, error)
	FindByPublicID(ctx context.Context, userID int64, publicID string) (*entities.DrugScan, error)
	FindByUserID(ctx context.Context, userID int64) ([]*entities.DrugScan, error)
	Save(ctx context.Context, scan *entities.DrugScan) error
	Update(ctx context.Context, scan *entities.DrugScan) error
	Delete(ctx context.Context, id int64) error
}
