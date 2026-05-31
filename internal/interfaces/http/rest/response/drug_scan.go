package response

import "time"

type DrugScanResponse struct {
	PublicID           string                      `json:"public_id"`
	DrugName           *string                     `json:"drug_name"`
	RegistrationNumber *string                     `json:"registration_number"`
	ExpiryDate         *time.Time                  `json:"expiry_date"`
	IsVerified         bool                        `json:"is_verified"`
	VerificationStatus *string                     `json:"verification_status"`
	Explanation        *string                     `json:"explanation"`
	ConfidenceScore    *float64                    `json:"confidence_score"`
	LotNumberValid     *bool                       `json:"lot_number_valid"`
	RegisteredMedicine *RegisteredMedicineResponse `json:"registered_medicine,omitempty"`
	CreatedAt          time.Time                   `json:"created_at"`
}

type RegisteredMedicineResponse struct {
	ID                 int64      `json:"id"`
	DrugName           string     `json:"drug_name"`
	RegistrationNumber string     `json:"registration_number"`
	Manufacturer       *string    `json:"manufacturer"`
	CountryCode        string     `json:"country_code"`
	Strength           *string    `json:"strength"`
	IngredientName     *string    `json:"ingredient_name"`
	CategoryName       *string    `json:"category_name"`
	FormName           *string    `json:"form_name"`
	RouteName          *string    `json:"route_name"`
	ApplicantName      *string    `json:"applicant_name"`
	RegisteredDate     *time.Time `json:"registered_date"`
	ExpiryDate         *time.Time `json:"expiry_date"`
	Status             bool       `json:"status"`
}
