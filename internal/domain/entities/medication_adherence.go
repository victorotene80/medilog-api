package entities

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type MedicationAdherenceLog struct {
	ID           int64
	MedicationID int64
	UserID       int64
	ScheduledAt  time.Time
	TakenAt      *time.Time
	Status       valueobjects.AdherenceStatus
	Note         *string
	CreatedAt    time.Time
}

func (l *MedicationAdherenceLog) IsTaken() bool {
	return l.Status == valueobjects.AdherenceStatusTaken
}

func (l *MedicationAdherenceLog) MarkTaken(now time.Time) {
	l.TakenAt = &now
	l.Status = valueobjects.AdherenceStatusTaken
}

func (l *MedicationAdherenceLog) MarkMissed() {
	l.Status = valueobjects.AdherenceStatusMissed
}

func (l *MedicationAdherenceLog) MarkSkipped(note string) {
	l.Status = valueobjects.AdherenceStatusSkipped
	l.Note = &note
}
