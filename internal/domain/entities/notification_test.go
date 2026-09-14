package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNotificationIsRead(t *testing.T) {
	now := time.Now().UTC()
	n := &Notification{ReadAt: &now}
	assert.True(t, n.IsRead())
}

func TestNotificationIsNotRead(t *testing.T) {
	n := &Notification{ReadAt: nil}
	assert.False(t, n.IsRead())
}

func TestNotificationIsSent(t *testing.T) {
	now := time.Now().UTC()
	n := &Notification{SentAt: &now}
	assert.True(t, n.IsSent())
}

func TestNotificationMarkRead(t *testing.T) {
	now := time.Now().UTC()
	n := &Notification{Status: "unread"}
	n.MarkRead(now)
	assert.Equal(t, &now, n.ReadAt)
	assert.Equal(t, "read", n.Status)
}
