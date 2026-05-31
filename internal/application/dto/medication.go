package dto

import "time"

type MedicationTimeDTO struct {
	ID        int64  `json:"id"`
	TimeValue string `json:"time_value"` // "HH:MM"
}

type MedicationDTO struct {
	ID                 int64               `json:"id"`
	PublicID           string              `json:"public_id"`
	UserID             int64               `json:"user_id"`
	Name               string              `json:"name"`
	DrugClass          *int                `json:"drug_class,omitempty"`
	Dosage             *string             `json:"dosage,omitempty"`
	Frequency          *string             `json:"frequency,omitempty"`
	WithFood           bool                `json:"with_food"`
	PrescribedBy       *string             `json:"prescribed_by,omitempty"`
	Facility           *string             `json:"facility,omitempty"`
	AddedVia           *string             `json:"added_via,omitempty"`
	RegistrationNumber *string             `json:"registration_number,omitempty"`
	RegCountryCode     *string             `json:"reg_country_code,omitempty"`
	IsVerified         bool                `json:"is_verified"`
	StartDate          *time.Time          `json:"start_date,omitempty"`
	EndDate            *time.Time          `json:"end_date,omitempty"`
	Notes              *string             `json:"notes,omitempty"`
	AdherenceRate      float64             `json:"adherence_rate"`
	IsCompleted        bool                `json:"is_completed"`
	CompletedDate      *time.Time          `json:"completed_date,omitempty"`
	Times              []MedicationTimeDTO `json:"times"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

type AdherenceLogDTO struct {
	ID           int64      `json:"id"`
	MedicationID int64      `json:"medication_id"`
	ScheduledAt  time.Time  `json:"scheduled_at"`
	TakenAt      *time.Time `json:"taken_at,omitempty"`
	Status       string     `json:"status"`
	Note         *string    `json:"note,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
