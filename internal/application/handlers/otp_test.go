package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeRecipient(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"mixed-case email", "  John.Doe@Example.COM ", "john.doe@example.com"},
		{"phone with spaces", "+234 803 987 6543", "+2348039876543"},
		{"phone with dashes and parens", "+1 (415) 555-0100", "+14155550100"},
		{"already normalized phone", "+2348039876543", "+2348039876543"},
		{"empty", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeRecipient(tt.input))
		})
	}
}
