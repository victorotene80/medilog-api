package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCountryCodeValid(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"US", "US"},
		{"us", "US"},
		{" Ng ", "NG"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cc, err := NewCountryCode(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, cc.String())
		})
	}
}

func TestNewCountryCodeInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"too short", "U"},
		{"too long", "USA"},
		{"numbers", "1A"},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCountryCode(tt.input)
			assert.Error(t, err)
		})
	}
}

func TestCountryCodeIsZero(t *testing.T) {
	cc := CountryCode{}
	assert.True(t, cc.IsZero())

	valid, _ := NewCountryCode("US")
	assert.False(t, valid.IsZero())
}

func TestCountryCodeEquals(t *testing.T) {
	a, _ := NewCountryCode("us")
	b, _ := NewCountryCode("US")
	c, _ := NewCountryCode("NG")

	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
