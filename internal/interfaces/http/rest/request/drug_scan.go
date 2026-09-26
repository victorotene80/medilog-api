package request

type VerifyDrugScanRequest struct {
	DrugName           *string `json:"drug_name"            validate:"omitempty,max=200"`
	RegistrationNumber *string `json:"registration_number"  validate:"omitempty,max=150"`
	CountryCode        string  `json:"country_code"         validate:"required,max=10"`
	// ExpiryDate is a string, not *time.Time: packs print "MM/YYYY", so the
	// client sends YYYY-MM, and a bad date must be a field error rather than
	// failing the whole decode as "Invalid JSON payload".
	ExpiryDate *string `json:"expiry_date" validate:"omitempty,max=40"`
}
