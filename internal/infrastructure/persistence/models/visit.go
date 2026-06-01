package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"gorm.io/gorm"
)

type VisitModel struct {
	ID             int64          `gorm:"column:id;primaryKey;autoIncrement"`
	PublicID       string         `gorm:"column:public_id;type:uuid;default:gen_random_uuid()"`
	UserID         int64          `gorm:"column:user_id;not null;index"`
	HospitalName   *string        `gorm:"column:hospital_name"`
	Diagnosis      *string        `gorm:"column:diagnosis"`
	VisitDate      time.Time      `gorm:"column:visit_date;not null;index"`
	Outcome        *string        `gorm:"column:outcome"`
	MedsCount      int            `gorm:"column:meds_count;not null;default:0"`
	Doctor         *string        `gorm:"column:doctor"`
	ChiefComplaint *string        `gorm:"column:chief_complaint"`
	Notes          *string        `gorm:"column:notes"`
	BloodPressure  *string        `gorm:"column:blood_pressure"`
	Temperature    *float64       `gorm:"column:temperature"`
	Weight         *float64       `gorm:"column:weight"`
	Pulse          *int           `gorm:"column:pulse"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (VisitModel) TableName() string { return "visits" }

func VisitToEntity(m *VisitModel) *entities.Visit {
	return &entities.Visit{
		ID:             m.ID,
		PublicID:       m.PublicID,
		UserID:         m.UserID,
		HospitalName:   m.HospitalName,
		Diagnosis:      m.Diagnosis,
		VisitDate:      m.VisitDate,
		Outcome:        m.Outcome,
		MedsCount:      m.MedsCount,
		Doctor:         m.Doctor,
		ChiefComplaint: m.ChiefComplaint,
		Notes:          m.Notes,
		BloodPressure:  m.BloodPressure,
		Temperature:    m.Temperature,
		Weight:         m.Weight,
		Pulse:          m.Pulse,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func VisitToModel(e *entities.Visit) *VisitModel {
	return &VisitModel{
		ID:             e.ID,
		PublicID:       e.PublicID,
		UserID:         e.UserID,
		HospitalName:   e.HospitalName,
		Diagnosis:      e.Diagnosis,
		VisitDate:      e.VisitDate,
		Outcome:        e.Outcome,
		MedsCount:      e.MedsCount,
		Doctor:         e.Doctor,
		ChiefComplaint: e.ChiefComplaint,
		Notes:          e.Notes,
		BloodPressure:  e.BloodPressure,
		Temperature:    e.Temperature,
		Weight:         e.Weight,
		Pulse:          e.Pulse,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}
