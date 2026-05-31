package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type FeedbackRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.Feedback, error)
	Save(ctx context.Context, f *entities.Feedback) error
	Update(ctx context.Context, f *entities.Feedback) error
}
