package valueobjects

import "errors"

type Sex int

const (
	SexMale   Sex = 1
	SexFemale Sex = 2
	SexOther  Sex = 3
)

func NewSex(v int) (Sex, error) {
	switch Sex(v) {
	case SexMale, SexFemale, SexOther:
		return Sex(v), nil
	}
	return 0, errors.New("invalid sex value: must be 1 (male), 2 (female), or 3 (other)")
}

func (s Sex) Int() int { return int(s) }

func (s Sex) String() string {
	switch s {
	case SexMale:
		return "male"
	case SexFemale:
		return "female"
	default:
		return "other"
	}
}
