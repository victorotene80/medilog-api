package response

import "time"

type MedicationTimeResponse struct {
	ID        int64  `json:"id"`
	TimeValue string `json:"time_value"`
}

type MedicationResponse struct {
	PublicID           string                   `json:"public_id"`
	Name               string                   `json:"name"`
	DrugClass          *int                     `json:"drug_class,omitempty"`
	Dosage             *string                  `json:"dosage,omitempty"`
	Frequency          *string                  `json:"frequency,omitempty"`
	WithFood           bool                     `json:"with_food"`
	PrescribedBy       *string                  `json:"prescribed_by,omitempty"`
	Facility           *string                  `json:"facility,omitempty"`
	AddedVia           *string                  `json:"added_via,omitempty"`
	RegistrationNumber *string                  `json:"registration_number,omitempty"`
	RegCountryCode     *string                  `json:"reg_country_code,omitempty"`
	IsVerified         bool                     `json:"is_verified"`
	StartDate          *time.Time               `json:"start_date,omitempty"`
	EndDate            *time.Time               `json:"end_date,omitempty"`
	Notes              *string                  `json:"notes,omitempty"`
	AdherenceRate      float64                  `json:"adherence_rate"`
	IsCompleted        bool                     `json:"is_completed"`
	CompletedDate      *time.Time               `json:"completed_date,omitempty"`
	Times              []MedicationTimeResponse `json:"times"`
	CreatedAt          time.Time                `json:"created_at"`
	UpdatedAt          time.Time                `json:"updated_at"`
}
