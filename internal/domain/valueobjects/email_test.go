package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmailValid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple", "user@example.com"},
		{"with dots", "first.last@example.com"},
		{"with plus", "user+tag@example.com"},
		{"with subdomain", "user@sub.example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := NewEmail(tt.input)
			assert.NoError(t, err)
			assert.False(t, e.IsZero())
		})
	}
}

func TestNewEmailInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"no at sign", "userexample.com"},
		{"no domain", "user@"},
		{"spaces", "user @example.com"},
		{"no tld", "user@example"},
		{"no tld dot", "user@example."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEmail(tt.input)
			assert.Error(t, err)
		})
	}
}

func TestNewEmailNormalized(t *testing.T) {
	e, err := NewEmail("  User@Example.COM  ")
	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", e.String())
}

func TestNewEmailEmpty(t *testing.T) {
	_, err := NewEmail("")
	assert.Error(t, err)
}
