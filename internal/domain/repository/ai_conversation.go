package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
)

type AIConversationRepository interface {
	FindByID(ctx context.Context, id int64) (*aggregates.AIConversationAggregate, error)
	FindByPublicID(ctx context.Context, userID int64, publicID string) (*aggregates.AIConversationAggregate, error)
	FindByUserID(ctx context.Context, userID int64) ([]*aggregates.AIConversationAggregate, error)
	FindActiveByUserID(ctx context.Context, userID int64) ([]*aggregates.AIConversationAggregate, error)
	Save(ctx context.Context, agg *aggregates.AIConversationAggregate) error
	Update(ctx context.Context, agg *aggregates.AIConversationAggregate) error
	Delete(ctx context.Context, id int64) error
}
