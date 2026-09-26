package request

type VerifyDrugScanRequest struct {
	DrugName           *string `json:"drug_name"            validate:"required_without=RegistrationNumber,omitempty,max=200"`
	RegistrationNumber *string `json:"registration_number"  validate:"required_without=DrugName,omitempty,max=150"`
	CountryCode        string  `json:"country_code"         validate:"required,max=10"`
}
