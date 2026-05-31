package models

import (
	"encoding/json"
	"time"
)

type AuditLogModel struct {
	ID          int64           `gorm:"column:id;primaryKey;autoIncrement"`
	Action      string          `gorm:"column:action;not null"`
	UserID      *string         `gorm:"column:user_id"`
	ActorID     *string         `gorm:"column:actor_id"`
	SessionID   *string         `gorm:"column:session_id"`
	IPAddress   *string         `gorm:"column:ip_address"`
	UserAgent   *string         `gorm:"column:user_agent"`
	CountryCode *string         `gorm:"column:country_code"`
	TargetID    *string         `gorm:"column:target_id"`
	Metadata    json.RawMessage `gorm:"column:metadata;type:jsonb"`
	Success     bool            `gorm:"column:success;not null;default:true"`
	OccurredAt  time.Time       `gorm:"column:occurred_at;not null"`
	CreatedAt   time.Time       `gorm:"column:created_at;autoCreateTime"`
}

func (AuditLogModel) TableName() string {
	return "audit_logs"
}
