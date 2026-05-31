package command

import "time"

type CreateVisitCommand struct {
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
}

type UpdateVisitCommand struct {
	UserID         int64
	PublicID       string
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
}

type DeleteVisitCommand struct {
	UserID   int64
	PublicID string
}
