package dto

import "time"

type VisitDTO struct {
	ID             int64     `json:"id"`
	PublicID       string    `json:"public_id"`
	UserID         int64     `json:"user_id"`
	HospitalName   *string   `json:"hospital_name,omitempty"`
	Diagnosis      *string   `json:"diagnosis,omitempty"`
	VisitDate      time.Time `json:"visit_date"`
	Outcome        *string   `json:"outcome,omitempty"`
	MedsCount      int       `json:"meds_count"`
	Doctor         *string   `json:"doctor,omitempty"`
	ChiefComplaint *string   `json:"chief_complaint,omitempty"`
	Notes          *string   `json:"notes,omitempty"`
	BloodPressure  *string   `json:"blood_pressure,omitempty"`
	Temperature    *float64  `json:"temperature,omitempty"`
	Weight         *float64  `json:"weight,omitempty"`
	Pulse          *int      `json:"pulse,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
