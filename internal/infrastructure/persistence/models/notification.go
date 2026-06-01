package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type NotificationModel struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement"`
	PublicID    string     `gorm:"column:public_id;type:uuid;default:gen_random_uuid()"`
	UserID      int64      `gorm:"column:user_id;not null;index"`
	Title       string     `gorm:"column:title;not null"`
	Body        string     `gorm:"column:body;not null"`
	Type        string     `gorm:"column:type;not null"`
	Channel     *string    `gorm:"column:channel"`
	Status      string     `gorm:"column:status;not null;default:unread;index"`
	ImageURL    *string    `gorm:"column:image_url"`
	ScheduledAt *time.Time `gorm:"column:scheduled_at"`
	SentAt      *time.Time `gorm:"column:sent_at"`
	ReadAt      *time.Time `gorm:"column:read_at"`
	Metadata    JSONMap    `gorm:"column:metadata;type:jsonb"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (NotificationModel) TableName() string { return "notifications" }

func NotificationToEntity(m *NotificationModel) *entities.Notification {
	return &entities.Notification{
		ID:          m.ID,
		PublicID:    m.PublicID,
		UserID:      m.UserID,
		Title:       m.Title,
		Body:        m.Body,
		Type:        m.Type,
		Channel:     m.Channel,
		Status:      m.Status,
		ImageURL:    m.ImageURL,
		ScheduledAt: m.ScheduledAt,
		SentAt:      m.SentAt,
		ReadAt:      m.ReadAt,
		Metadata:    map[string]any(m.Metadata),
		CreatedAt:   m.CreatedAt,
	}
}

func NotificationToModel(e *entities.Notification) *NotificationModel {
	return &NotificationModel{
		ID:          e.ID,
		PublicID:    e.PublicID,
		UserID:      e.UserID,
		Title:       e.Title,
		Body:        e.Body,
		Type:        e.Type,
		Channel:     e.Channel,
		Status:      e.Status,
		ImageURL:    e.ImageURL,
		ScheduledAt: e.ScheduledAt,
		SentAt:      e.SentAt,
		ReadAt:      e.ReadAt,
		Metadata:    JSONMap(e.Metadata),
		CreatedAt:   e.CreatedAt,
	}
}
