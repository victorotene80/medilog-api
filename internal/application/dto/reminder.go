package dto

// ReminderRunResultDTO summarises one scheduler tick. Skipped counts slots that
// already existed, which in steady state is nearly all of them.
type ReminderRunResultDTO struct {
	Scanned            int  `json:"scanned"`
	MedicationCreated  int  `json:"medication_created"`
	AppointmentCreated int  `json:"appointment_created"`
	Skipped            int  `json:"skipped"`
	LockNotAcquired    bool `json:"lock_not_acquired"`
}
