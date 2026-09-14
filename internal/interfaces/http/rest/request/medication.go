package request

import "time"

// TimeValue is a wall-clock HH:MM in the user's timezone. The format is
// enforced here so a malformed value is a 400 at the edge rather than a parse
// error deep in the handler, which surfaced as a 500.
type MedicationTimeRequest struct {
	TimeValue string `json:"time_value" validate:"required,len=5,datetime=15:04"`
}

type CreateMedicationRequest struct {
	Name               string                  `json:"name"               validate:"required,max=200"`
	DrugClass          *int                    `json:"drug_class"`
	Dosage             *string                 `json:"dosage"             validate:"omitempty,max=100"`
	Frequency          *string                 `json:"frequency"          validate:"omitempty,max=100"`
	WithFood           bool                    `json:"with_food"`
	PrescribedBy       *string                 `json:"prescribed_by"      validate:"omitempty,max=150"`
	Facility           *string                 `json:"facility"           validate:"omitempty,max=200"`
	AddedVia           *string                 `json:"added_via"          validate:"omitempty,max=50"`
	RegistrationNumber *string                 `json:"registration_number" validate:"omitempty,max=150"`
	RegCountryCode     *string                 `json:"reg_country_code"   validate:"omitempty,max=3"`
	StartDate          *time.Time              `json:"start_date"`
	EndDate            *time.Time              `json:"end_date"`
	Notes              *string                 `json:"notes"`
	Times              []MedicationTimeRequest `json:"times"             validate:"omitempty,dive"`
}

type UpdateMedicationRequest struct {
	Name         string                  `json:"name"          validate:"required,max=200"`
	DrugClass    *int                    `json:"drug_class"`
	Dosage       *string                 `json:"dosage"        validate:"omitempty,max=100"`
	Frequency    *string                 `json:"frequency"     validate:"omitempty,max=100"`
	WithFood     bool                    `json:"with_food"`
	PrescribedBy *string                 `json:"prescribed_by" validate:"omitempty,max=150"`
	Facility     *string                 `json:"facility"      validate:"omitempty,max=200"`
	AddedVia     *string                 `json:"added_via"     validate:"omitempty,max=50"`
	StartDate    *time.Time              `json:"start_date"`
	EndDate      *time.Time              `json:"end_date"`
	Notes        *string                 `json:"notes"`
	Times        []MedicationTimeRequest `json:"times"         validate:"omitempty,dive"`
}

type LogAdherenceRequest struct {
	ScheduledAt time.Time `json:"scheduled_at" validate:"required"`
	Status      int       `json:"status"       validate:"required,min=1,max=3"`
	Note        *string   `json:"note"         validate:"omitempty,max=500"`
}
