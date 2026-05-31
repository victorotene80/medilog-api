package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type MedicationTimeModel struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	MedicationID int64     `gorm:"column:medication_id;not null;index"`
	TimeValue    time.Time `gorm:"column:time_value;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (MedicationTimeModel) TableName() string { return "medication_times" }

func MedicationTimeToEntity(m *MedicationTimeModel) *entities.MedicationTime {
	return &entities.MedicationTime{
		ID:           m.ID,
		MedicationID: m.MedicationID,
		TimeValue:    m.TimeValue,
		CreatedAt:    m.CreatedAt,
	}
}

func MedicationTimeToModel(e *entities.MedicationTime) *MedicationTimeModel {
	return &MedicationTimeModel{
		ID:           e.ID,
		MedicationID: e.MedicationID,
		TimeValue:    e.TimeValue,
		CreatedAt:    e.CreatedAt,
	}
}
