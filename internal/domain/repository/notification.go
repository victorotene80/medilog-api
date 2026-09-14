package repository

import (
	"context"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type NotificationRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.Notification, error)
	// FindByPublicID scopes the lookup to userID in SQL rather than filtering
	// after the fetch, so a caller can never observe another user's row.
	FindByPublicID(ctx context.Context, userID int64, publicID string) (*entities.Notification, error)
	FindByUserID(ctx context.Context, userID int64) ([]*entities.Notification, error)
	FindUnreadByUserID(ctx context.Context, userID int64) ([]*entities.Notification, error)
	Save(ctx context.Context, n *entities.Notification) error
	// SaveIfNotExists inserts a server-generated notification, doing nothing if
	// one already exists with the same (user_id, dedupe_key). It reports whether
	// a row was actually created. This is what makes the reminder scheduler safe
	// to re-run over an overlapping window, and safe to run on several replicas.
	SaveIfNotExists(ctx context.Context, n *entities.Notification) (created bool, err error)
	Update(ctx context.Context, n *entities.Notification) error
	MarkAllReadByUserID(ctx context.Context, userID int64, now time.Time) error
	CountUnreadByUserID(ctx context.Context, userID int64) (int64, error)
}
