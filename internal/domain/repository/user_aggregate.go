package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
)

type UserAggregateRepository interface {
	FindByID(ctx context.Context, id int64) (*aggregates.UserAggregate, error)
	FindByPublicID(ctx context.Context, publicID string) (*aggregates.UserAggregate, error)
	FindByEmail(ctx context.Context, email string) (*aggregates.UserAggregate, error)
	FindByPhone(ctx context.Context, phone string) (*aggregates.UserAggregate, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByPhone(ctx context.Context, phone string) (bool, error)
	Save(ctx context.Context, agg *aggregates.UserAggregate) error
	Update(ctx context.Context, agg *aggregates.UserAggregate) error
	Delete(ctx context.Context, id int64) error
	DeleteByPublicID(ctx context.Context, publicID string) error
}
