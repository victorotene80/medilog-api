package valueobjects

import "errors"

type BloodType struct {
	value string
}

var validBloodTypes = map[string]struct{}{
	"A+": {}, "A-": {},
	"B+": {}, "B-": {},
	"O+": {}, "O-": {},
	"AB+": {}, "AB-": {},
}

func NewBloodType(raw string) (BloodType, error) {
	if _, ok := validBloodTypes[raw]; !ok {
		return BloodType{}, errors.New("invalid blood type: must be one of A+, A-, B+, B-, O+, O-, AB+, AB-")
	}
	return BloodType{value: raw}, nil
}

func (b BloodType) String() string          { return b.value }
func (b BloodType) Equals(o BloodType) bool { return b.value == o.value }
