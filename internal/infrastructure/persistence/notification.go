package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) FindByID(ctx context.Context, id int64) (*entities.Notification, error) {
	var m models.NotificationModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.NotificationToEntity(&m), nil
}

func (r *NotificationRepository) FindByUserID(ctx context.Context, userID int64) ([]*entities.Notification, error) {
	var ms []models.NotificationModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&ms).Error; err != nil {
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
	if err := r.db.WithContext(ctx).
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
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	n.ID = m.ID
	return nil
}

func (r *NotificationRepository) Update(ctx context.Context, n *entities.Notification) error {
	return r.db.WithContext(ctx).Save(models.NotificationToModel(n)).Error
}

func (r *NotificationRepository) MarkAllReadByUserID(ctx context.Context, userID int64, now time.Time) error {
	return r.db.WithContext(ctx).Model(&models.NotificationModel{}).
		Where("user_id = ? AND status = 'unread'", userID).
		Updates(map[string]any{"status": "read", "read_at": now}).Error
}
