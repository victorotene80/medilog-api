package readmodel

import "time"

// DueMedicationCandidate is one (medication, time-of-day) pair belonging to a
// user whose medication reminders are enabled.
//
// TimeOfDay is a wall-clock "HH:MM" already extracted in UTC. medication_times
// rows are written by time.Parse("15:04", ...), which yields a zero date at a
// UTC offset — so the stored instant carries no real date and must be read with
// an explicit AT TIME ZONE 'UTC' rather than a bare ::time cast, which would
// shift under a non-UTC session TimeZone.
type DueMedicationCandidate struct {
	UserID           int64
	Timezone         string
	MedicationID     int64
	MedicationTimeID int64
	MedicationName   string
	Dosage           *string
	Frequency        *string
	StartDate        *time.Time
	EndDate          *time.Time
	TimeOfDay        string
}

// DueAppointmentCandidate is an upcoming visit for a user who wants
// appointment reminders.
type DueAppointmentCandidate struct {
	UserID       int64
	Timezone     string
	VisitID      int64
	HospitalName *string
	Doctor       *string
	VisitDate    time.Time
}
