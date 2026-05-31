package command

import "time"

type MedicationTimeInput struct {
	TimeValue string `json:"time_value"` // "HH:MM" 24-hour
}

type CreateMedicationCommand struct {
	UserID             int64
	Name               string
	DrugClass          *int
	Dosage             *string
	Frequency          *string
	WithFood           bool
	PrescribedBy       *string
	Facility           *string
	AddedVia           *string
	RegistrationNumber *string
	RegCountryCode     *string
	StartDate          *time.Time
	EndDate            *time.Time
	Notes              *string
	Times              []MedicationTimeInput
}

type UpdateMedicationCommand struct {
	UserID       int64
	PublicID     string
	Name         string
	DrugClass    *int
	Dosage       *string
	Frequency    *string
	WithFood     bool
	PrescribedBy *string
	Facility     *string
	AddedVia     *string
	StartDate    *time.Time
	EndDate      *time.Time
	Notes        *string
	Times        []MedicationTimeInput
}

type CompleteMedicationCommand struct {
	UserID   int64
	PublicID string
}

type DeleteMedicationCommand struct {
	UserID   int64
	PublicID string
}

type LogMedicationAdherenceCommand struct {
	UserID       int64
	MedicationID int64
	ScheduledAt  time.Time
	Status       int
	Note         *string
}
