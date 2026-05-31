package dto

import "time"

type AuditAction string

const (
	AuditActionLoginSuccess    AuditAction = "login.success"
	AuditActionLoginFailed     AuditAction = "login.failed"
	AuditActionLogout          AuditAction = "logout"
	AuditActionPasswordChanged AuditAction = "password.changed"
	AuditActionPasswordReset   AuditAction = "password.reset"
	AuditActionOTPRequested    AuditAction = "otp.requested"
	AuditActionOTPVerified     AuditAction = "otp.verified"
)

type AuditRecord struct {
	Action      AuditAction
	UserID      *string
	ActorID     *string
	SessionID   *string
	IPAddress   *string
	UserAgent   *string
	CountryCode *string
	TargetID    *string
	Metadata    map[string]any
	Success     bool
	OccurredAt  time.Time
}
