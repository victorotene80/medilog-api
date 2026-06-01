package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"gorm.io/gorm"
)

type FunFactModel struct {
	ID                int64          `gorm:"column:id;primaryKey;autoIncrement"`
	Title             string         `gorm:"column:title;not null"`
	Text              string         `gorm:"column:text;not null"`
	Category          *string        `gorm:"column:category"`
	TargetCountryCode *string        `gorm:"column:target_country_code"`
	TargetAgeMin      *int           `gorm:"column:target_age_min"`
	TargetAgeMax      *int           `gorm:"column:target_age_max"`
	AllergyCategory   *int           `gorm:"column:allergy_category"`
	IsActive          bool           `gorm:"column:is_active;not null;default:true"`
	CreatedAt         time.Time      `gorm:"column:created_at;autoCreateTime"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (FunFactModel) TableName() string { return "fun_facts" }

func FunFactToEntity(m *FunFactModel) *entities.FunFact {
	return &entities.FunFact{
		ID:                m.ID,
		Title:             m.Title,
		Text:              m.Text,
		Category:          m.Category,
		TargetCountryCode: m.TargetCountryCode,
		TargetAgeMin:      m.TargetAgeMin,
		TargetAgeMax:      m.TargetAgeMax,
		AllergyCategory:   m.AllergyCategory,
		IsActive:          m.IsActive,
		CreatedAt:         m.CreatedAt,
	}
}

func FunFactToModel(e *entities.FunFact) *FunFactModel {
	return &FunFactModel{
		ID:                e.ID,
		Title:             e.Title,
		Text:              e.Text,
		Category:          e.Category,
		TargetCountryCode: e.TargetCountryCode,
		TargetAgeMin:      e.TargetAgeMin,
		TargetAgeMax:      e.TargetAgeMax,
		AllergyCategory:   e.AllergyCategory,
		IsActive:          e.IsActive,
		CreatedAt:         e.CreatedAt,
	}
}
