package valueobjects

import "errors"

type AllergySeverity int16

const (
	AllergySeverityMild            AllergySeverity = 1
	AllergySeverityModerate        AllergySeverity = 2
	AllergySeveritySevere          AllergySeverity = 3
	AllergySeverityLifeThreatening AllergySeverity = 4
	AllergySeverityUnknown         AllergySeverity = 5
)

func NewAllergySeverity(v int16) (AllergySeverity, error) {
	s := AllergySeverity(v)

	switch s {
	case AllergySeverityMild,
		AllergySeverityModerate,
		AllergySeveritySevere,
		AllergySeverityLifeThreatening,
		AllergySeverityUnknown:
		return s, nil
	default:
		return 0, errors.New("invalid allergy severity")
	}
}

func (s AllergySeverity) Int16() int16 {
	return int16(s)
}

func (s AllergySeverity) String() string {
	switch s {
	case AllergySeverityMild:
		return "Mild"
	case AllergySeverityModerate:
		return "Moderate"
	case AllergySeveritySevere:
		return "Severe"
	case AllergySeverityLifeThreatening:
		return "Life threatening"
	case AllergySeverityUnknown:
		return "Unknown"
	default:
		return "Invalid"
	}
}
