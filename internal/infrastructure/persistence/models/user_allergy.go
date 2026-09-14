package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
	"gorm.io/gorm"
)

type UserAllergyModel struct {
	ID          int64          `gorm:"column:id;primaryKey;autoIncrement"`
	PublicID    string         `gorm:"column:public_id;type:uuid;default:gen_random_uuid()"`
	UserID      int64          `gorm:"column:user_id;not null"`
	AllergyID   *int64         `gorm:"column:allergy_id"`
	Name        string         `gorm:"column:name;not null"`
	Description *string        `gorm:"column:description"`
	Severity    *int16         `gorm:"column:severity"` // smallint → *int16, never a value object
	Category    int            `gorm:"column:category;not null"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (UserAllergyModel) TableName() string {
	return "user_allergies"
}

func UserAllergyToEntity(m *UserAllergyModel) *entities.UserAllergy {
	if m == nil {
		return nil
	}

	category := valueobjects.AllergyCategory(m.Category)

	var severity *valueobjects.AllergySeverity
	if m.Severity != nil {
		v := valueobjects.AllergySeverity(*m.Severity)
		severity = &v
	}

	return &entities.UserAllergy{
		ID:          m.ID,
		PublicID:    m.PublicID,
		UserID:      m.UserID,
		AllergyID:   m.AllergyID,
		Name:        m.Name,
		Description: m.Description,
		Severity:    severity,
		Category:    category,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func UserAllergyEntityToModel(e *entities.UserAllergy) *UserAllergyModel {
	if e == nil {
		return nil
	}

	var severity *int16
	if e.Severity != nil {
		v := int16(*e.Severity)
		severity = &v
	}

	return &UserAllergyModel{
		ID:          e.ID,
		PublicID:    e.PublicID,
		UserID:      e.UserID,
		AllergyID:   e.AllergyID,
		Name:        e.Name,
		Description: e.Description,
		Severity:    severity,
		Category:    int(e.Category),
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}
