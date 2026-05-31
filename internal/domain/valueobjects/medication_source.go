package valueobjects

import "errors"

type MedicationSource string

const (
	MedicationSourceManual MedicationSource = "manual"
	MedicationSourceScan   MedicationSource = "scan"
	MedicationSourceVisit  MedicationSource = "visit"
	MedicationSourceImport MedicationSource = "import"
)

func NewMedicationSource(raw string) (MedicationSource, error) {
	s := MedicationSource(raw)
	switch s {
	case MedicationSourceManual, MedicationSourceScan,
		MedicationSourceVisit, MedicationSourceImport:
		return s, nil
	}
	return "", errors.New("invalid medication source")
}

func (s MedicationSource) String() string { return string(s) }
