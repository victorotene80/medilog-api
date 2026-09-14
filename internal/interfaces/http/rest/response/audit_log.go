package response

import "time"

type AuditLogResponse struct {
	ID          string         `json:"id"`
	Action      string         `json:"action"`
	UserID      *string        `json:"user_id,omitempty"`
	ActorID     *string        `json:"actor_id,omitempty"`
	SessionID   *string        `json:"session_id,omitempty"`
	IPAddress   *string        `json:"ip_address,omitempty"`
	UserAgent   *string        `json:"user_agent,omitempty"`
	CountryCode *string        `json:"country_code,omitempty"`
	TargetID    *string        `json:"target_id,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Success     bool           `json:"success"`
	OccurredAt  time.Time      `json:"occurred_at"`
	CreatedAt   time.Time      `json:"created_at"`
}

type ListAuditLogsResponse struct {
	Logs   []*AuditLogResponse `json:"logs"`
	Total  int64               `json:"total"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
}
