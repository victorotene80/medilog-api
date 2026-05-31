package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type RegisteredMedicineModel struct {
	ID                 int64      `gorm:"column:id;primaryKey;autoIncrement"`
	RegulatoryBodyID   int64      `gorm:"column:regulatory_body_id;not null"`
	CountryCode        string     `gorm:"column:country_code;not null"`
	DrugName           string     `gorm:"column:drug_name;not null"`
	RegistrationNumber string     `gorm:"column:registration_number;not null"`
	Barcode            *string    `gorm:"column:barcode"`
	Manufacturer       *string    `gorm:"column:manufacturer"`
	RegisteredDate     *time.Time `gorm:"column:registered_date"`
	ExpiryDate         *time.Time `gorm:"column:expiry_date"`
	Status             bool       `gorm:"column:status;not null"`
	SourceProductID    *int64     `gorm:"column:source_product_id"`
	Strength           *string    `gorm:"column:strength"`
	IngredientName     *string    `gorm:"column:ingredient_name"`
	Synonym            *string    `gorm:"column:synonym"`
	CategoryName       *string    `gorm:"column:category_name"`
	FormName           *string    `gorm:"column:form_name"`
	RouteName          *string    `gorm:"column:route_name"`
	ApplicantName      *string    `gorm:"column:applicant_name"`
	SourcePayload      *string    `gorm:"column:source_payload;type:jsonb"`
	CreatedAt          time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (RegisteredMedicineModel) TableName() string { return "registered_medicines" }

func RegisteredMedicineToEntity(m *RegisteredMedicineModel) *entities.RegisteredMedicine {
	return &entities.RegisteredMedicine{
		ID:                 m.ID,
		RegulatoryBodyID:   m.RegulatoryBodyID,
		CountryCode:        m.CountryCode,
		DrugName:           m.DrugName,
		RegistrationNumber: m.RegistrationNumber,
		Barcode:            m.Barcode,
		Manufacturer:       m.Manufacturer,
		RegisteredDate:     m.RegisteredDate,
		ExpiryDate:         m.ExpiryDate,
		Status:             m.Status,
		SourceProductID:    m.SourceProductID,
		Strength:           m.Strength,
		IngredientName:     m.IngredientName,
		Synonym:            m.Synonym,
		CategoryName:       m.CategoryName,
		FormName:           m.FormName,
		RouteName:          m.RouteName,
		ApplicantName:      m.ApplicantName,
		SourcePayload:      m.SourcePayload,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}

func RegisteredMedicineToModel(e *entities.RegisteredMedicine) *RegisteredMedicineModel {
	return &RegisteredMedicineModel{
		ID:                 e.ID,
		RegulatoryBodyID:   e.RegulatoryBodyID,
		CountryCode:        e.CountryCode,
		DrugName:           e.DrugName,
		RegistrationNumber: e.RegistrationNumber,
		Barcode:            e.Barcode,
		Manufacturer:       e.Manufacturer,
		RegisteredDate:     e.RegisteredDate,
		ExpiryDate:         e.ExpiryDate,
		Status:             e.Status,
		SourceProductID:    e.SourceProductID,
		Strength:           e.Strength,
		IngredientName:     e.IngredientName,
		Synonym:            e.Synonym,
		CategoryName:       e.CategoryName,
		FormName:           e.FormName,
		RouteName:          e.RouteName,
		ApplicantName:      e.ApplicantName,
		SourcePayload:      e.SourcePayload,
		CreatedAt:          e.CreatedAt,
		UpdatedAt:          e.UpdatedAt,
	}
}
