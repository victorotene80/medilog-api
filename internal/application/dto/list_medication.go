package dto

type ListMedicationsResponseDTO struct {
	Medications []*MedicationDTO `json:"medications"`
	NextCursor  *string          `json:"next_cursor,omitempty"`
}
