package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type MedicationModel struct {
	ID                 int64      `gorm:"column:id;primaryKey;autoIncrement"`
	PublicID           string     `gorm:"column:public_id;type:uuid;default:gen_random_uuid()"`
	UserID             int64      `gorm:"column:user_id;not null;index"`
	Name               string     `gorm:"column:name;not null"`
	DrugClass          *int       `gorm:"column:drug_class"`
	Dosage             *string    `gorm:"column:dosage"`
	Frequency          *string    `gorm:"column:frequency"`
	WithFood           bool       `gorm:"column:with_food;not null;default:false"`
	PrescribedBy       *string    `gorm:"column:prescribed_by"`
	Facility           *string    `gorm:"column:facility"`
	AddedVia           *string    `gorm:"column:added_via"`
	RegistrationNumber *string    `gorm:"column:registration_number"`
	RegCountryCode     *string    `gorm:"column:reg_country_code"`
	IsVerified         bool       `gorm:"column:is_verified;not null;default:false"`
	StartDate          *time.Time `gorm:"column:start_date"`
	EndDate            *time.Time `gorm:"column:end_date"`
	Notes              *string    `gorm:"column:notes"`
	AdherenceCount     int        `gorm:"column:adherence_count;not null;default:0"`
	TotalDoses         int        `gorm:"column:total_doses;not null;default:0"`
	IsCompleted        bool       `gorm:"column:is_completed;not null;default:false"`
	CompletedDate      *time.Time `gorm:"column:completed_date"`
	CreatedAt          time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (MedicationModel) TableName() string { return "medications" }

func MedicationToEntity(m *MedicationModel) (*entities.Medication, error) {
	entity := &entities.Medication{
		ID:                 m.ID,
		PublicID:           m.PublicID,
		UserID:             m.UserID,
		Name:               m.Name,
		DrugClass:          m.DrugClass,
		Dosage:             m.Dosage,
		WithFood:           m.WithFood,
		PrescribedBy:       m.PrescribedBy,
		Facility:           m.Facility,
		RegistrationNumber: m.RegistrationNumber,
		RegCountryCode:     m.RegCountryCode,
		IsVerified:         m.IsVerified,
		StartDate:          m.StartDate,
		EndDate:            m.EndDate,
		Notes:              m.Notes,
		AdherenceCount:     m.AdherenceCount,
		TotalDoses:         m.TotalDoses,
		IsCompleted:        m.IsCompleted,
		CompletedDate:      m.CompletedDate,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
	if m.Frequency != nil {
		freq, err := valueobjects.NewMedicationFrequency(m.Frequency)
		if err != nil {
			return nil, err
		}
		entity.Frequency = freq
	}
	if m.AddedVia != nil {
		src, err := valueobjects.NewMedicationSource(*m.AddedVia)
		if err != nil {
			return nil, err
		}
		entity.AddedVia = &src
	}
	return entity, nil
}

func MedicationToModel(e *entities.Medication) *MedicationModel {
	m := &MedicationModel{
		ID:                 e.ID,
		PublicID:           e.PublicID,
		UserID:             e.UserID,
		Name:               e.Name,
		DrugClass:          e.DrugClass,
		Dosage:             e.Dosage,
		WithFood:           e.WithFood,
		PrescribedBy:       e.PrescribedBy,
		Facility:           e.Facility,
		RegistrationNumber: e.RegistrationNumber,
		RegCountryCode:     e.RegCountryCode,
		IsVerified:         e.IsVerified,
		StartDate:          e.StartDate,
		EndDate:            e.EndDate,
		Notes:              e.Notes,
		AdherenceCount:     e.AdherenceCount,
		TotalDoses:         e.TotalDoses,
		IsCompleted:        e.IsCompleted,
		CompletedDate:      e.CompletedDate,
		CreatedAt:          e.CreatedAt,
		UpdatedAt:          e.UpdatedAt,
	}
	if e.Frequency != nil {
		s := e.Frequency.String()
		m.Frequency = &s
	}
	if e.AddedVia != nil {
		s := e.AddedVia.String()
		m.AddedVia = &s
	}
	return m
}
