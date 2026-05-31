package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type DrugScanModel struct {
	ID                   int64      `gorm:"column:id;primaryKey;autoIncrement"`
	PublicID             string     `gorm:"column:public_id;type:uuid;default:gen_random_uuid()"`
	UserID               int64      `gorm:"column:user_id;not null;index"`
	DrugName             *string    `gorm:"column:drug_name"`
	RegistrationNumber   *string    `gorm:"column:registration_number"`
	ExpiryDate           *time.Time `gorm:"column:expiry_date"`
	RegulatoryBodyID     *int64     `gorm:"column:regulatory_body_id"`
	RegisteredMedicineID *int64     `gorm:"column:registered_medicine_id"`
	IsVerified           bool       `gorm:"column:is_verified;not null;default:false"`
	ConfidenceScore      *float64   `gorm:"column:confidence_score"`
	LotNumberValid       *bool      `gorm:"column:lot_number_valid"`
	VerificationStatus   *string    `gorm:"column:verification_status"`
	Explanation          *string    `gorm:"column:explanation"`
	CreatedAt            time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (DrugScanModel) TableName() string { return "drug_scans" }

func DrugScanToEntity(m *DrugScanModel) *entities.DrugScan {
	return &entities.DrugScan{
		ID:                   m.ID,
		PublicID:             m.PublicID,
		UserID:               m.UserID,
		DrugName:             m.DrugName,
		RegistrationNumber:   m.RegistrationNumber,
		ExpiryDate:           m.ExpiryDate,
		RegulatoryBodyID:     m.RegulatoryBodyID,
		RegisteredMedicineID: m.RegisteredMedicineID,
		IsVerified:           m.IsVerified,
		ConfidenceScore:      m.ConfidenceScore,
		LotNumberValid:       m.LotNumberValid,
		VerificationStatus:   m.VerificationStatus,
		Explanation:          m.Explanation,
		CreatedAt:            m.CreatedAt,
	}
}

func DrugScanToModel(e *entities.DrugScan) *DrugScanModel {
	return &DrugScanModel{
		ID:                   e.ID,
		PublicID:             e.PublicID,
		UserID:               e.UserID,
		DrugName:             e.DrugName,
		RegistrationNumber:   e.RegistrationNumber,
		ExpiryDate:           e.ExpiryDate,
		RegulatoryBodyID:     e.RegulatoryBodyID,
		RegisteredMedicineID: e.RegisteredMedicineID,
		IsVerified:           e.IsVerified,
		ConfidenceScore:      e.ConfidenceScore,
		LotNumberValid:       e.LotNumberValid,
		VerificationStatus:   e.VerificationStatus,
		Explanation:          e.Explanation,
		CreatedAt:            e.CreatedAt,
	}
}
