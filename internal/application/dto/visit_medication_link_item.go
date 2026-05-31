package dto

import "time"

type VisitMedicationLinkItemDTO struct {
	ID           string    `json:"id"`
	VisitID      string    `json:"visit_id"`
	MedicationID string    `json:"medication_id"`
	CreatedAt    time.Time `json:"created_at"`
}
