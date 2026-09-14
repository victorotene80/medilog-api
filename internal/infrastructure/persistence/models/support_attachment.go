package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"gorm.io/gorm"
)

type SupportAttachmentModel struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement"`
	MessageID int64          `gorm:"column:message_id;not null;index"`
	FileURL   string         `gorm:"column:file_url;not null"`
	FileName  *string        `gorm:"column:file_name"`
	FileType  *string        `gorm:"column:file_type"`
	FileSize  *int64         `gorm:"column:file_size"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (SupportAttachmentModel) TableName() string { return "support_attachments" }

func SupportAttachmentToEntity(m *SupportAttachmentModel) *entities.SupportAttachment {
	return &entities.SupportAttachment{
		ID:        m.ID,
		MessageID: m.MessageID,
		FileURL:   m.FileURL,
		FileName:  m.FileName,
		FileType:  m.FileType,
		FileSize:  m.FileSize,
		CreatedAt: m.CreatedAt,
	}
}

func SupportAttachmentToModel(e *entities.SupportAttachment) *SupportAttachmentModel {
	return &SupportAttachmentModel{
		ID:        e.ID,
		MessageID: e.MessageID,
		FileURL:   e.FileURL,
		FileName:  e.FileName,
		FileType:  e.FileType,
		FileSize:  e.FileSize,
		CreatedAt: e.CreatedAt,
	}
}
