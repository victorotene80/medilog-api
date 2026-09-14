package response

import "time"

// AIQuotaResponse drives the question-limit UI: the header counter, the limit
// modal, and whether the composer is disabled.
//
// When IsPro is true the account is uncapped and Used/Total/Remaining are not
// meaningful limits.
type AIQuotaResponse struct {
	Used      int       `json:"used"`
	Total     int       `json:"total"`
	Remaining int       `json:"remaining"`
	IsPro     bool      `json:"is_pro"`
	ResetsAt  time.Time `json:"resets_at"`
}
