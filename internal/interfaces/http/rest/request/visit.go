package request

import "time"

type CreateVisitRequest struct {
	HospitalName   *string   `json:"hospital_name"             validate:"omitempty,min=1,max=200"`
	Diagnosis      *string   `json:"diagnosis"                 validate:"omitempty,min=1,max=500"`
	VisitDate      time.Time `json:"visit_date"                validate:"required"`
	Outcome        *string   `json:"outcome"                   validate:"omitempty,min=1,max=100"`
	MedsCount      *int      `json:"meds_count,omitempty"      validate:"omitempty,min=0"`
	Doctor         *string   `json:"doctor,omitempty"          validate:"omitempty,max=150"`
	ChiefComplaint *string   `json:"chief_complaint,omitempty" validate:"omitempty,max=500"`
	Notes          *string   `json:"notes,omitempty"           validate:"omitempty,max=1000"`
	BloodPressure  *string   `json:"blood_pressure,omitempty"  validate:"omitempty,max=50"`
	Temperature    *float64  `json:"temperature,omitempty"     validate:"omitempty,min=30,max=45"`
	Weight         *float64  `json:"weight,omitempty"          validate:"omitempty,min=1,max=500"`
	Pulse          *int      `json:"pulse,omitempty"           validate:"omitempty,min=30,max=250"`
}

type UpdateVisitRequest CreateVisitRequest
