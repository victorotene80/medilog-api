package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type UserProfileRepository interface {
	FindByUserID(ctx context.Context, userID int64) (*entities.UserProfile, error)
	Save(ctx context.Context, profile *entities.UserProfile) error
	Update(ctx context.Context, profile *entities.UserProfile) error
}
