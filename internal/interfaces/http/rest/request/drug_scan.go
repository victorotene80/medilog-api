package request

import "time"

type VerifyDrugScanRequest struct {
	DrugName           *string    `json:"drug_name"            validate:"omitempty,max=200"`
	RegistrationNumber *string    `json:"registration_number"  validate:"omitempty,max=150"`
	CountryCode        string     `json:"country_code"         validate:"required,max=10"`
	ExpiryDate         *time.Time `json:"expiry_date"`
}
