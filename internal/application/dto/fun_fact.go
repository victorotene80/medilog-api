package dto

import "time"

type FunFactDTO struct {
	ID                int64     `json:"id"`
	Title             *string   `json:"title,omitempty"`
	Text              string    `json:"text"`
	Category          *string   `json:"category,omitempty"`
	TargetCountryCode *string   `json:"target_country_code,omitempty"`
	TargetAgeMin      *int      `json:"target_age_min,omitempty"`
	TargetAgeMax      *int      `json:"target_age_max,omitempty"`
	AllergyCategory   *int      `json:"allergy_category,omitempty"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
}
