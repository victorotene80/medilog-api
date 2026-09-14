package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBloodTypeValid(t *testing.T) {
	valid := []string{"A+", "A-", "B+", "B-", "O+", "O-", "AB+", "AB-"}
	for _, v := range valid {
		t.Run(v, func(t *testing.T) {
			bt, err := NewBloodType(v)
			assert.NoError(t, err)
			assert.Equal(t, v, bt.String())
		})
	}
}

func TestNewBloodTypeInvalid(t *testing.T) {
	invalid := []string{"C+", "ABO+", "o+", "", "1+"}
	for _, v := range invalid {
		t.Run(v, func(t *testing.T) {
			_, err := NewBloodType(v)
			assert.Error(t, err)
		})
	}
}
