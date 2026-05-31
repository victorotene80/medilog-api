package repository

import (
	"context"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
)

type MedicationRepository interface {
	FindByID(ctx context.Context, id int64) (*aggregates.MedicationAggregate, error)
	FindByPublicID(ctx context.Context, userID int64, publicID string) (*aggregates.MedicationAggregate, error)
	FindByUserID(ctx context.Context, userID int64) ([]*aggregates.MedicationAggregate, error)
	FindActiveByUserID(ctx context.Context, userID int64, now time.Time) ([]*aggregates.MedicationAggregate, error)
	Save(ctx context.Context, agg *aggregates.MedicationAggregate) error
	Update(ctx context.Context, agg *aggregates.MedicationAggregate) error
	Delete(ctx context.Context, id int64) error
}
