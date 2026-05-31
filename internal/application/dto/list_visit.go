package dto

type ListVisitsResponseDTO struct {
	Visits     []VisitDTO `json:"visits"`
	NextCursor *string    `json:"next_cursor,omitempty"`
}
