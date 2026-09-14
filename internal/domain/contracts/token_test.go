package contracts

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToken_Fields(t *testing.T) {
	expiresAt := time.Date(2025, 1, 15, 13, 0, 0, 0, time.UTC)

	token := Token{
		Value:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		ExpiresAt: expiresAt,
	}

	assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", token.Value)
	assert.Equal(t, expiresAt, token.ExpiresAt)
}

func TestTokenPair_Fields(t *testing.T) {
	accessExpiry := time.Date(2025, 1, 15, 13, 0, 0, 0, time.UTC)
	refreshExpiry := time.Date(2025, 1, 22, 13, 0, 0, 0, time.UTC)

	pair := TokenPair{
		AccessToken: Token{
			Value:     "access-token-value",
			ExpiresAt: accessExpiry,
		},
		RefreshToken: Token{
			Value:     "refresh-token-value",
			ExpiresAt: refreshExpiry,
		},
	}

	assert.Equal(t, "access-token-value", pair.AccessToken.Value)
	assert.Equal(t, accessExpiry, pair.AccessToken.ExpiresAt)
	assert.Equal(t, "refresh-token-value", pair.RefreshToken.Value)
	assert.Equal(t, refreshExpiry, pair.RefreshToken.ExpiresAt)
}

func TestToken_EmptyValue(t *testing.T) {
	token := Token{}

	assert.Empty(t, token.Value)
	assert.True(t, token.ExpiresAt.IsZero())
}

func TestTokenPair_DifferentExpiries(t *testing.T) {
	now := time.Now().UTC()

	pair := TokenPair{
		AccessToken: Token{
			Value:     "access",
			ExpiresAt: now.Add(15 * time.Minute),
		},
		RefreshToken: Token{
			Value:     "refresh",
			ExpiresAt: now.Add(7 * 24 * time.Hour),
		},
	}

	assert.True(t, pair.RefreshToken.ExpiresAt.After(pair.AccessToken.ExpiresAt))
}
