package repository

import (
	"context"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type RefreshTokenRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.RefreshToken, error)
	FindActiveByUserID(ctx context.Context, userID int64) ([]*entities.RefreshToken, error)
	Save(ctx context.Context, token *entities.RefreshToken) error
	Update(ctx context.Context, token *entities.RefreshToken) error
	RevokeIfActive(ctx context.Context, id int64, now time.Time, replacedByHash *string) (bool, error)
	RevokeAllForUser(ctx context.Context, userID int64, now time.Time) error
	DeleteExpired(ctx context.Context, before time.Time) error
}
