package request

import "time"

type MedicationTimeRequest struct {
	TimeValue string `json:"time_value" validate:"required"`
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
	Times              []MedicationTimeRequest `json:"times"`
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
	Times        []MedicationTimeRequest `json:"times"`
}

type LogAdherenceRequest struct {
	ScheduledAt time.Time `json:"scheduled_at" validate:"required"`
	Status      int       `json:"status"       validate:"required,min=1,max=3"`
	Note        *string   `json:"note"         validate:"omitempty,max=500"`
}

/*type MedicationTimeRequest struct {
	Label string     `json:"label" validate:"required,min=1,max=50"`
	Value *time.Time `json:"value,omitempty"`
}

type CreateMedicationRequest struct {
	Name         string                  `json:"name"                   validate:"required,min=1,max=300"`
	DrugClass    string                  `json:"drug_class"             validate:"required,min=1,max=200"`
	Dosage       string                  `json:"dosage"                 validate:"required,min=1,max=100"`
	Frequency    string                  `json:"frequency"              validate:"required,min=1,max=100"`
	WithFood     bool                    `json:"with_food"`
	IsVerified   bool                    `json:"is_verified"`
	PrescribedBy *string                 `json:"prescribed_by,omitempty"  validate:"omitempty,max=200"`
	Facility     *string                 `json:"facility,omitempty"       validate:"omitempty,max=300"`
	AddedVia     *string                 `json:"added_via,omitempty"      validate:"omitempty"`
	NafdacNumber *string                 `json:"nafdac_number,omitempty"  validate:"omitempty,max=50"`
	StartDate    *time.Time              `json:"start_date,omitempty"`
	EndDate      *time.Time              `json:"end_date,omitempty"`
	Notes        *string                 `json:"notes,omitempty"          validate:"omitempty,max=1000"`
	IconURL      *string                 `json:"icon_url,omitempty"       validate:"omitempty,url"`
	Times        []MedicationTimeRequest `json:"times,omitempty"          validate:"omitempty,dive"`
}

type UpdateMedicationRequest struct {
	Name           string                  `json:"name"                     validate:"required,min=1,max=300"`
	DrugClass      string                  `json:"drug_class"               validate:"required,min=1,max=200"`
	Dosage         string                  `json:"dosage"                   validate:"required,min=1,max=100"`
	Frequency      string                  `json:"frequency"                validate:"required,min=1,max=100"`
	WithFood       bool                    `json:"with_food"`
	PrescribedBy   *string                 `json:"prescribed_by,omitempty"  validate:"omitempty,max=200"`
	Facility       *string                 `json:"facility,omitempty"       validate:"omitempty,max=300"`
	MedicineNumber *string                 `json:"medicine_number,omitempty" validate:"omitempty,max=50"`
	StartDate      *time.Time              `json:"start_date,omitempty"`
	EndDate        *time.Time              `json:"end_date,omitempty"`
	Notes          *string                 `json:"notes,omitempty"          validate:"omitempty,max=1000"`
	IconURL        *string                 `json:"icon_url,omitempty"       validate:"omitempty,url"`
	Times          []MedicationTimeRequest `json:"times,omitempty"          validate:"omitempty,dive"`
}*/
