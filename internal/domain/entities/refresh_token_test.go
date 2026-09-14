package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRefreshTokenIsExpired(t *testing.T) {
	past := time.Now().UTC().Add(-1 * time.Hour)
	rt := &RefreshToken{ExpiresAt: past}
	assert.True(t, rt.IsExpired(time.Now().UTC()))
}

func TestRefreshTokenIsNotExpired(t *testing.T) {
	future := time.Now().UTC().Add(1 * time.Hour)
	rt := &RefreshToken{ExpiresAt: future}
	assert.False(t, rt.IsExpired(time.Now().UTC()))
}

func TestRefreshTokenIsRevoked(t *testing.T) {
	now := time.Now().UTC()
	rt := &RefreshToken{RevokedAt: &now}
	assert.True(t, rt.IsRevoked())
}

func TestRefreshTokenIsNotRevoked(t *testing.T) {
	rt := &RefreshToken{RevokedAt: nil}
	assert.False(t, rt.IsRevoked())
}

func TestRefreshTokenIsValid(t *testing.T) {
	now := time.Now().UTC()
	future := now.Add(1 * time.Hour)

	valid := &RefreshToken{ExpiresAt: future, RevokedAt: nil}
	assert.True(t, valid.IsValid(now))

	expired := &RefreshToken{ExpiresAt: now.Add(-1 * time.Hour), RevokedAt: nil}
	assert.False(t, expired.IsValid(now))

	revoked := &RefreshToken{ExpiresAt: future, RevokedAt: &now}
	assert.False(t, revoked.IsValid(now))
}

func TestRefreshTokenRevoke(t *testing.T) {
	now := time.Now().UTC()
	replacement := "new-hash"
	rt := &RefreshToken{}
	rt.Revoke(now, &replacement)
	assert.Equal(t, &now, rt.RevokedAt)
	assert.Equal(t, &replacement, rt.ReplacedByTokenHash)
}
