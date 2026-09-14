package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type UserProfileRepository interface {
	FindByUserID(ctx context.Context, userID int64) (*entities.UserProfile, error)
	Save(ctx context.Context, profile *entities.UserProfile) error
	Update(ctx context.Context, profile *entities.UserProfile) error
	// ConsumeAIQuestion increments the counter in a single conditional UPDATE
	// and reports whether it succeeded. A read-modify-write here would let two
	// concurrent sends both observe used=9 and both write used=10, handing out
	// one more question than the plan allows.
	ConsumeAIQuestion(ctx context.Context, userID int64) (bool, error)
}
