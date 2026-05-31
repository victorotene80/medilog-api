package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type MedicationAdherenceLogModel struct {
	ID           int64      `gorm:"column:id;primaryKey;autoIncrement"`
	MedicationID int64      `gorm:"column:medication_id;not null"`
	UserID       int64      `gorm:"column:user_id;not null"`
	ScheduledAt  time.Time  `gorm:"column:scheduled_at;not null;index:idx_adherence_user_scheduled,priority:2"`
	TakenAt      *time.Time `gorm:"column:taken_at"`
	Status       int        `gorm:"column:status;not null"`
	Note         *string    `gorm:"column:note"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (MedicationAdherenceLogModel) TableName() string { return "medication_adherence_logs" }

func MedicationAdherenceLogToEntity(m *MedicationAdherenceLogModel) (*entities.MedicationAdherenceLog, error) {
	status, err := valueobjects.NewAdherenceStatus(m.Status)
	if err != nil {
		return nil, err
	}
	return &entities.MedicationAdherenceLog{
		ID:           m.ID,
		MedicationID: m.MedicationID,
		UserID:       m.UserID,
		ScheduledAt:  m.ScheduledAt,
		TakenAt:      m.TakenAt,
		Status:       status,
		Note:         m.Note,
		CreatedAt:    m.CreatedAt,
	}, nil
}

func MedicationAdherenceLogToModel(e *entities.MedicationAdherenceLog) *MedicationAdherenceLogModel {
	return &MedicationAdherenceLogModel{
		ID:           e.ID,
		MedicationID: e.MedicationID,
		UserID:       e.UserID,
		ScheduledAt:  e.ScheduledAt,
		TakenAt:      e.TakenAt,
		Status:       e.Status.Int(),
		Note:         e.Note,
		CreatedAt:    e.CreatedAt,
	}
}
