package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type FeedbackModel struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID      *int64    `gorm:"column:user_id"`
	Rating      *int      `gorm:"column:rating"`
	Title       *string   `gorm:"column:title"`
	Message     string    `gorm:"column:message;not null"`
	AppVersion  *string   `gorm:"column:app_version"`
	Platform    *string   `gorm:"column:platform"`
	DeviceModel *string   `gorm:"column:device_model"`
	Status      string    `gorm:"column:status;not null;default:open"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (FeedbackModel) TableName() string { return "feedback" }

func FeedbackToEntity(m *FeedbackModel) *entities.Feedback {
	return &entities.Feedback{
		ID:          m.ID,
		UserID:      m.UserID,
		Rating:      m.Rating,
		Title:       m.Title,
		Message:     m.Message,
		AppVersion:  m.AppVersion,
		Platform:    m.Platform,
		DeviceModel: m.DeviceModel,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt,
	}
}

func FeedbackToModel(e *entities.Feedback) *FeedbackModel {
	return &FeedbackModel{
		ID:          e.ID,
		UserID:      e.UserID,
		Rating:      e.Rating,
		Title:       e.Title,
		Message:     e.Message,
		AppVersion:  e.AppVersion,
		Platform:    e.Platform,
		DeviceModel: e.DeviceModel,
		Status:      e.Status,
		CreatedAt:   e.CreatedAt,
	}
}
