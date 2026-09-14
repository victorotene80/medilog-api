package dto

import "time"

// AIQuotaDTO is the shape the client's question-limit UI renders: the header
// counter, the limit modal, and whether the composer is disabled.
type AIQuotaDTO struct {
	Used      int       `json:"used"`
	Total     int       `json:"total"`
	Remaining int       `json:"remaining"`
	IsPro     bool      `json:"is_pro"`
	ResetsAt  time.Time `json:"resets_at"`
}
