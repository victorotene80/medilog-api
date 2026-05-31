package repository

import (
	"context"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type DashboardRepository interface {
	GetByUserID(ctx context.Context, userID int64, now time.Time) (*entities.Dashboard, error)
}
