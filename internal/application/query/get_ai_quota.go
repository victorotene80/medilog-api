package query

// GetAIQuotaQuery reads the caller's current AI question allowance.
type GetAIQuotaQuery struct {
	UserID int64
}
