package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMedicationAdherenceRate(t *testing.T) {
	m := &Medication{TotalDoses: 0, AdherenceCount: 0}
	assert.Equal(t, float64(0), m.AdherenceRate())
}

func TestMedicationAdherenceRateCalculation(t *testing.T) {
	m := &Medication{TotalDoses: 10, AdherenceCount: 7}
	assert.Equal(t, 70.0, m.AdherenceRate())
}

func TestMedicationIsActive(t *testing.T) {
	now := time.Now().UTC()
	completed := &Medication{IsCompleted: true}
	assert.False(t, completed.IsActive(now))
}

func TestMedicationIsActiveEndDate(t *testing.T) {
	now := time.Now().UTC()
	past := now.Add(-24 * time.Hour)
	m := &Medication{EndDate: &past}
	assert.False(t, m.IsActive(now))
}

func TestMedicationIsActiveActive(t *testing.T) {
	now := time.Now().UTC()
	m := &Medication{IsCompleted: false, EndDate: nil}
	assert.True(t, m.IsActive(now))
}

func TestMedicationComplete(t *testing.T) {
	now := time.Now().UTC()
	m := &Medication{}
	m.Complete(now)
	assert.True(t, m.IsCompleted)
	assert.Equal(t, &now, m.CompletedDate)
}

func TestMedicationRecordDose(t *testing.T) {
	m := &Medication{}
	m.RecordDose(true)
	assert.Equal(t, 1, m.TotalDoses)
	assert.Equal(t, 1, m.AdherenceCount)

	m.RecordDose(false)
	assert.Equal(t, 2, m.TotalDoses)
	assert.Equal(t, 1, m.AdherenceCount)
}
