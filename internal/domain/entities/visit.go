package entities

import "time"

type Visit struct {
	ID             int64
	PublicID       string
	UserID         int64
	HospitalName   *string
	Diagnosis      *string
	VisitDate      time.Time
	Outcome        *string
	MedsCount      int
	Doctor         *string
	ChiefComplaint *string
	Notes          *string
	BloodPressure  *string
	Temperature    *float64
	Weight         *float64
	Pulse          *int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (v *Visit) HasVitals() bool {
	return v.BloodPressure != nil || v.Temperature != nil ||
		v.Weight != nil || v.Pulse != nil
}
