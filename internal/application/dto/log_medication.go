package dto

import "time"

type LogMedicationDoseResponseDTO struct {
	MedicationID   string    `json:"medication_id"`
	Status         string    `json:"status"`
	AdherenceCount int       `json:"adherence_count"`
	LoggedAt       time.Time `json:"logged_at"`
}
