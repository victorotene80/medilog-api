package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTimezoneValid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"iana name", "Africa/Lagos", "Africa/Lagos"},
		{"utc", "UTC", "UTC"},
		{"surrounding whitespace is trimmed", "  Europe/London  ", "Europe/London"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tz, err := NewTimezone(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, tz)
		})
	}
}

func TestNewTimezoneInvalid(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"empty", "", ErrTimezoneRequired},
		{"whitespace only", "   ", ErrTimezoneRequired},
		{"not a zone", "Mars/Olympus", ErrTimezoneInvalid},
		// A bare offset is what a client sends when it confuses an IANA name
		// with a UTC offset. It has to be rejected: the scheduler needs a zone
		// to apply DST rules with, which an offset cannot express.
		{"utc offset", "+01:00", ErrTimezoneInvalid},
		{"abbreviation", "WAT", ErrTimezoneInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tz, err := NewTimezone(tt.input)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Empty(t, tz)
		})
	}
}
