package repository

import (
	"context"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type NotificationRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.Notification, error)
	FindByUserID(ctx context.Context, userID int64) ([]*entities.Notification, error)
	FindUnreadByUserID(ctx context.Context, userID int64) ([]*entities.Notification, error)
	Save(ctx context.Context, n *entities.Notification) error
	Update(ctx context.Context, n *entities.Notification) error
	MarkAllReadByUserID(ctx context.Context, userID int64, now time.Time) error
}
