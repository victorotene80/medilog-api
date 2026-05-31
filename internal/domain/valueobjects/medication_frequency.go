package valueobjects

import (
	"errors"
	"strings"
)

type MedicationFrequency string

const (
	FrequencyOnce     MedicationFrequency = "once"
	FrequencyDaily    MedicationFrequency = "daily"
	FrequencyTwiceDay MedicationFrequency = "twice_daily"
	FrequencyThreeDay MedicationFrequency = "three_times_daily"
	FrequencyWeekly   MedicationFrequency = "weekly"
	FrequencyAsNeeded MedicationFrequency = "as_needed"
)

func NewMedicationFrequency(raw *string) (*MedicationFrequency, error) {
	if raw == nil {
		return nil, nil
	}

	value := strings.TrimSpace(*raw)
	if value == "" {
		return nil, nil
	}

	f := MedicationFrequency(value)

	switch f {
	case FrequencyOnce, FrequencyDaily, FrequencyTwiceDay,
		FrequencyThreeDay, FrequencyWeekly, FrequencyAsNeeded:
		return &f, nil
	default:
		return nil, errors.New("invalid medication frequency")
	}
}

func (f MedicationFrequency) String() string {
	return string(f)
}
