package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPhoneValid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"with plus", "+1234567890"},
		{"without plus", "1234567890"},
		{"long number", "+123456789012345"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPhone(tt.input)
			assert.NoError(t, err)
			assert.False(t, p.IsZero())
		})
	}
}

func TestNewPhoneInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"starts with zero", "+0123456789"},
		{"too short", "+123"},
		{"letters", "+abc1234567"},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPhone(tt.input)
			assert.Error(t, err)
		})
	}
}

func TestNewPhoneTrimmed(t *testing.T) {
	p, err := NewPhone("  +1234567890  ")
	assert.NoError(t, err)
	assert.Equal(t, "+1234567890", p.String())
}
