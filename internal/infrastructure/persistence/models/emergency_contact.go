package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"gorm.io/gorm"
)

type EmergencyContactModel struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement"`
	PublicID     string         `gorm:"column:public_id;type:uuid;default:gen_random_uuid()"`
	UserID       int64          `gorm:"column:user_id;not null"`
	Name         string         `gorm:"column:name;not null"`
	Relationship string         `gorm:"column:relationship;not null"`
	Phone        string         `gorm:"column:phone;not null"`
	IsPrimary    bool           `gorm:"column:is_primary;not null;default:false"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (EmergencyContactModel) TableName() string {
	return "emergency_contacts"
}

func EmergencyContactToEntity(m *EmergencyContactModel) *entities.EmergencyContact {
	if m == nil {
		return nil
	}

	return &entities.EmergencyContact{
		ID:           m.ID,
		PublicID:     m.PublicID,
		UserID:       m.UserID,
		Name:         m.Name,
		Relationship: m.Relationship,
		Phone:        m.Phone,
		IsPrimary:    m.IsPrimary,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func EmergencyContactEntityToModel(e *entities.EmergencyContact) *EmergencyContactModel {
	if e == nil {
		return nil
	}

	return &EmergencyContactModel{
		ID:           e.ID,
		PublicID:     e.PublicID,
		UserID:       e.UserID,
		Name:         e.Name,
		Relationship: e.Relationship,
		Phone:        e.Phone,
		IsPrimary:    e.IsPrimary,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}
