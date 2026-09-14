package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMedicationFrequencyValid(t *testing.T) {
	tests := []struct {
		input string
		want  MedicationFrequency
	}{
		{"once", FrequencyOnce},
		{"daily", FrequencyDaily},
		{"twice_daily", FrequencyTwiceDay},
		{"three_times_daily", FrequencyThreeDay},
		{"weekly", FrequencyWeekly},
		{"as_needed", FrequencyAsNeeded},
	}
	for _, tt := range tests {
		t.Run(string(tt.want), func(t *testing.T) {
			f, err := NewMedicationFrequency(&tt.input)
			assert.NoError(t, err)
			assert.NotNil(t, f)
			assert.Equal(t, tt.want, *f)
		})
	}
}

func TestNewMedicationFrequencyNil(t *testing.T) {
	f, err := NewMedicationFrequency(nil)
	assert.NoError(t, err)
	assert.Nil(t, f)
}

func TestNewMedicationFrequencyInvalid(t *testing.T) {
	input := "every_hour"
	f, err := NewMedicationFrequency(&input)
	assert.Error(t, err)
	assert.Nil(t, f)
}
