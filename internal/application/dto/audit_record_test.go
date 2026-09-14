package dto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuditRecord_Fields(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	userID := "user-123"
	actorID := "actor-456"
	sessionID := "session-789"
	ip := "192.168.1.1"
	ua := "Mozilla/5.0"
	country := "US"
	targetID := "target-012"

	record := AuditRecord{
		Action:      AuditActionLoginSuccess,
		UserID:      &userID,
		ActorID:     &actorID,
		SessionID:   &sessionID,
		IPAddress:   &ip,
		UserAgent:   &ua,
		CountryCode: &country,
		TargetID:    &targetID,
		Success:     true,
		OccurredAt:  now,
	}

	assert.Equal(t, AuditActionLoginSuccess, record.Action)
	assert.Equal(t, &userID, record.UserID)
	assert.Equal(t, &actorID, record.ActorID)
	assert.Equal(t, &sessionID, record.SessionID)
	assert.Equal(t, &ip, record.IPAddress)
	assert.Equal(t, &ua, record.UserAgent)
	assert.Equal(t, &country, record.CountryCode)
	assert.Equal(t, &targetID, record.TargetID)
	assert.True(t, record.Success)
	assert.Equal(t, now, record.OccurredAt)
}

func TestAuditRecord_WithMetadata(t *testing.T) {
	record := AuditRecord{
		Action: AuditActionLoginFailed,
		Metadata: map[string]any{
			"reason":  "invalid_password",
			"attempts": 3,
		},
		Success:    false,
		OccurredAt: time.Now().UTC(),
	}

	assert.Equal(t, AuditActionLoginFailed, record.Action)
	assert.False(t, record.Success)
	assert.NotNil(t, record.Metadata)
	assert.Equal(t, "invalid_password", record.Metadata["reason"])
	assert.Equal(t, 3, record.Metadata["attempts"])
}

func TestAuditRecord_NilOptionalFields(t *testing.T) {
	record := AuditRecord{
		Action:     AuditActionLogout,
		Success:    true,
		OccurredAt: time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
	}

	assert.Nil(t, record.UserID)
	assert.Nil(t, record.ActorID)
	assert.Nil(t, record.SessionID)
	assert.Nil(t, record.IPAddress)
	assert.Nil(t, record.UserAgent)
	assert.Nil(t, record.CountryCode)
	assert.Nil(t, record.TargetID)
	assert.Nil(t, record.Metadata)
}
