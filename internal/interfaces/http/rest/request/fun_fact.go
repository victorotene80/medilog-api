package request

type CreateFunFactRequest struct {
	Title             string  `json:"title"                validate:"required,min=1,max=200"`
	Text              string  `json:"text"                 validate:"required,min=1"`
	Category          *string `json:"category"             validate:"omitempty,max=100"`
	TargetCountryCode *string `json:"target_country_code"  validate:"omitempty,max=10"`
	TargetAgeMin      *int    `json:"target_age_min"       validate:"omitempty,min=0"`
	TargetAgeMax      *int    `json:"target_age_max"       validate:"omitempty,min=0"`
	AllergyCategory   *int    `json:"allergy_category"     validate:"omitempty,min=1,max=5"`
	IsActive          *bool   `json:"is_active"`
}

type UpdateFunFactRequest struct {
	Title             string  `json:"title"                validate:"required,min=1,max=200"`
	Text              string  `json:"text"                 validate:"required,min=1"`
	Category          *string `json:"category"             validate:"omitempty,max=100"`
	TargetCountryCode *string `json:"target_country_code"  validate:"omitempty,max=10"`
	TargetAgeMin      *int    `json:"target_age_min"       validate:"omitempty,min=0"`
	TargetAgeMax      *int    `json:"target_age_max"       validate:"omitempty,min=0"`
	AllergyCategory   *int    `json:"allergy_category"     validate:"omitempty,min=1,max=5"`
	IsActive          *bool   `json:"is_active"`
}
