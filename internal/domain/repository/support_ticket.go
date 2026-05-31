package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
)

type SupportTicketRepository interface {
	FindByID(ctx context.Context, id int64) (*aggregates.SupportTicketAggregate, error)
	FindByUserID(ctx context.Context, userID int64) ([]*aggregates.SupportTicketAggregate, error)
	FindOpenByUserID(ctx context.Context, userID int64) ([]*aggregates.SupportTicketAggregate, error)
	Save(ctx context.Context, agg *aggregates.SupportTicketAggregate) error
	Update(ctx context.Context, agg *aggregates.SupportTicketAggregate) error
}
