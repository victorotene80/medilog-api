package request

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Without `dive` the validator does not descend into slice elements, so the
// `required` on TimeValue never fired and "9am" reached the handler's parser —
// which returns a bare error, so StatusFrom mapped it to 500 instead of 4xx.
func TestCreateMedicationRequest_TimesAreValidatedPerElement(t *testing.T) {
	v := validator.New()

	cases := []struct {
		name    string
		time    string
		wantErr bool
	}{
		{"valid midnight", "00:00", false},
		{"valid morning", "09:00", false},
		{"valid evening", "20:30", false},
		{"non-numeric", "9am", true},
		{"empty", "", true},
		{"out of range hour", "25:00", true},
		{"unpadded", "9:00", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := CreateMedicationRequest{
				Name:  "Metformin",
				Times: []MedicationTimeRequest{{TimeValue: tc.time}},
			}

			err := v.Struct(req)
			if tc.wantErr {
				require.Error(t, err, "%q must be rejected at the edge", tc.time)
				return
			}
			assert.NoError(t, err, "%q is a legitimate dose time", tc.time)
		})
	}
}

// An absent times[] is legitimate — a medication need not have a schedule.
func TestCreateMedicationRequest_TimesOptional(t *testing.T) {
	assert.NoError(t, validator.New().Struct(CreateMedicationRequest{Name: "Metformin"}))
}
