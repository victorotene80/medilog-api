package valueobjects

import "errors"

type AdherenceStatus int

const (
	AdherenceStatusPending AdherenceStatus = 0
	AdherenceStatusTaken   AdherenceStatus = 1
	AdherenceStatusMissed  AdherenceStatus = 2
	AdherenceStatusSkipped AdherenceStatus = 3
)

func NewAdherenceStatus(v int) (AdherenceStatus, error) {
	s := AdherenceStatus(v)
	switch s {
	case AdherenceStatusPending, AdherenceStatusTaken,
		AdherenceStatusMissed, AdherenceStatusSkipped:
		return s, nil
	}
	return 0, errors.New("invalid adherence status")
}

func (s AdherenceStatus) Int() int { return int(s) }
