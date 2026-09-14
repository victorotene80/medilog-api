package persistence

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ repository.NotificationRepository = (*NotificationRepository)(nil)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) FindByID(ctx context.Context, id int64) (*entities.Notification, error) {
	var m models.NotificationModel
	if err := conn(ctx, r.db).Where("deleted_at IS NULL").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.NotificationToEntity(&m), nil
}

// FindByPublicID scopes to user_id in the WHERE clause so another user's
// notification is indistinguishable from a missing one (404, never 403).
func (r *NotificationRepository) FindByPublicID(
	ctx context.Context,
	userID int64,
	publicID string,
) (*entities.Notification, error) {
	if userID <= 0 {
		return nil, errors.New("user id is required")
	}

	publicID = strings.TrimSpace(publicID)
	if publicID == "" {
		return nil, errors.New("public id is required")
	}

	var m models.NotificationModel

	err := conn(ctx, r.db).
		Where("deleted_at IS NULL").
		Where("user_id = ? AND public_id = ?", userID, publicID).
		First(&m).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return models.NotificationToEntity(&m), nil
}

func (r *NotificationRepository) FindByUserID(ctx context.Context, userID int64) ([]*entities.Notification, error) {
	var ms []models.NotificationModel
	if err := conn(ctx, r.db).Where("deleted_at IS NULL").Where("user_id = ?", userID).Order("created_at DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	result := make([]*entities.Notification, len(ms))
	for i, m := range ms {
		m := m
		result[i] = models.NotificationToEntity(&m)
	}
	return result, nil
}

func (r *NotificationRepository) FindUnreadByUserID(ctx context.Context, userID int64) ([]*entities.Notification, error) {
	var ms []models.NotificationModel
	if err := conn(ctx, r.db).
		Where("deleted_at IS NULL").
		Where("user_id = ? AND status = 'unread'", userID).
		Order("created_at DESC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	result := make([]*entities.Notification, len(ms))
	for i, m := range ms {
		m := m
		result[i] = models.NotificationToEntity(&m)
	}
	return result, nil
}

func (r *NotificationRepository) Save(ctx context.Context, n *entities.Notification) error {
	m := models.NotificationToModel(n)
	if err := conn(ctx, r.db).Create(m).Error; err != nil {
		return err
	}
	n.ID = m.ID
	return nil
}

// SaveIfNotExists relies on the ux_notifications_user_dedupe unique index.
// Two replicas racing on the same slot both issue this insert; exactly one row
// results and the loser simply reports created=false.
func (r *NotificationRepository) SaveIfNotExists(
	ctx context.Context,
	n *entities.Notification,
) (bool, error) {
	if n == nil {
		return false, errors.New("notification is required")
	}

	if n.DedupeKey == nil || strings.TrimSpace(*n.DedupeKey) == "" {
		return false, errors.New("dedupe key is required for deduplicated inserts")
	}

	m := models.NotificationToModel(n)

	result := conn(ctx, r.db).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "dedupe_key"}},
			DoNothing: true,
		}).
		Create(m)

	if result.Error != nil {
		return false, result.Error
	}

	if result.RowsAffected == 0 {
		return false, nil
	}

	n.ID = m.ID
	n.PublicID = m.PublicID

	return true, nil
}

// Update writes the mutable columns only.
//
// It deliberately does not use Save: NotificationToModel does not carry
// DeletedAt, so a full Save would write deleted_at = NULL and resurrect a
// notification the user had dismissed.
func (r *NotificationRepository) Update(ctx context.Context, n *entities.Notification) error {
	if n == nil {
		return errors.New("notification is required")
	}

	if n.ID <= 0 {
		return errors.New("notification id is required")
	}

	result := conn(ctx, r.db).
		Model(&models.NotificationModel{}).
		Where("id = ?", n.ID).
		Updates(map[string]any{
			"title":        n.Title,
			"body":         n.Body,
			"type":         n.Type,
			"channel":      n.Channel,
			"status":       n.Status,
			"image_url":    n.ImageURL,
			"scheduled_at": n.ScheduledAt,
			"sent_at":      n.SentAt,
			"read_at":      n.ReadAt,
			"metadata":     models.JSONMap(n.Metadata),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *NotificationRepository) CountUnreadByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64

	err := conn(ctx, r.db).
		Model(&models.NotificationModel{}).
		Where("deleted_at IS NULL").
		Where("user_id = ? AND status = 'unread'", userID).
		Count(&count).Error

	return count, err
}

func (r *NotificationRepository) MarkAllReadByUserID(ctx context.Context, userID int64, now time.Time) error {
	return conn(ctx, r.db).Model(&models.NotificationModel{}).
		Where("user_id = ? AND status = 'unread'", userID).
		Updates(map[string]any{"status": "read", "read_at": now}).Error
}
