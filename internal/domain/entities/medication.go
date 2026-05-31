package entities

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type MedicationTime struct {
	ID           int64
	MedicationID int64
	TimeValue    time.Time // only the time portion is meaningful
	CreatedAt    time.Time
}
type Medication struct {
	ID                 int64
	PublicID           string
	UserID             int64
	Name               string
	DrugClass          *int
	Dosage             *string
	Frequency          *valueobjects.MedicationFrequency
	WithFood           bool
	PrescribedBy       *string
	Facility           *string
	AddedVia           *valueobjects.MedicationSource
	RegistrationNumber *string
	RegCountryCode     *string
	IsVerified         bool
	StartDate          *time.Time
	EndDate            *time.Time
	Notes              *string
	AdherenceCount     int
	TotalDoses         int
	IsCompleted        bool
	CompletedDate      *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (m *Medication) AdherenceRate() float64 {
	if m.TotalDoses == 0 {
		return 0
	}
	return float64(m.AdherenceCount) / float64(m.TotalDoses) * 100
}

func (m *Medication) IsActive(now time.Time) bool {
	if m.IsCompleted {
		return false
	}
	if m.EndDate != nil && now.After(*m.EndDate) {
		return false
	}
	return true
}

func (m *Medication) Complete(now time.Time) {
	m.IsCompleted = true
	m.CompletedDate = &now
	m.UpdatedAt = now
}

func (m *Medication) RecordDose(taken bool) {
	m.TotalDoses++
	if taken {
		m.AdherenceCount++
	}
}
