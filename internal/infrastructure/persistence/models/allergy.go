package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type AllergyModel struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string    `gorm:"column:name;not null"`
	Category    int       `gorm:"column:category;not null"`
	Description *string   `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (AllergyModel) TableName() string { return "allergies" }

func AllergyToEntity(m AllergyModel) *entities.Allergy {
	category, _ := valueobjects.NewAllergyCategory(m.Category)

	return &entities.Allergy{
		ID:          m.ID,
		Name:        m.Name,
		Category:    category,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
	}
}

func AllergyToModel(e entities.Allergy) *AllergyModel {
	return &AllergyModel{
		ID:          e.ID,
		Name:        e.Name,
		Category:    e.Category.Int(),
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
	}
}
