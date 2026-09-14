package request

type AddAllergyRequest struct {
	Name        string  `json:"name"        validate:"required,min=1,max=120"`
	Category    int     `json:"category"    validate:"required,min=1,max=5"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

type UpdateAllergyRequest struct {
	Name        string  `json:"name"        validate:"required,min=1,max=120"`
	Category    int     `json:"category"    validate:"required,min=1,max=5"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

type UserAllergyItem struct {
	AllergyID   *int64  `json:"allergy_id"  validate:"omitempty,min=1"`
	Name        string  `json:"name"        validate:"required,min=1,max=120"`
	Description *string `json:"description" validate:"omitempty,max=500"`
	// Severity: 1=Mild 2=Moderate 3=Severe 4=Life-threatening 5=Unknown
	Severity *int16 `json:"severity"    validate:"omitempty,min=1,max=5"`
	// Category: 1=Food 2=Drug 3=Environmental 4=Contact 5=Other
	Category int `json:"category"    validate:"required,min=1,max=5"`
}

type AddUserAllergiesRequest struct {
	Allergies []UserAllergyItem `json:"allergies" validate:"required,min=1,dive"`
}

// UpdateUserAllergyRequest fully replaces one of the user's allergy records.
// There is no allergy_id: editing a record detaches it from the reference
// catalogue, so the caller supplies the values directly.
type UpdateUserAllergyRequest struct {
	Name        string  `json:"name"        validate:"required,min=1,max=120"`
	Description *string `json:"description" validate:"omitempty,max=500"`
	// Severity: 1=Mild 2=Moderate 3=Severe 4=Life-threatening 5=Unknown
	Severity *int16 `json:"severity"    validate:"omitempty,min=1,max=5"`
	// Category: 1=Food 2=Drug 3=Environmental 4=Contact 5=Other
	Category int `json:"category"    validate:"required,min=1,max=5"`
}
