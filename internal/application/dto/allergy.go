package dto

import "time"

type AllergyDTO struct {
	ID          int64
	Name        string
	Category    int
	CategoryStr string
	Description *string
}

type UserAllergyDTO struct {
	PublicID    string
	AllergyID   *int64
	Name        string
	Description *string
	Severity    *int16
	SeverityStr *string
	Category    int
	CategoryStr string
	IsCustom    bool
	CreatedAt   time.Time
}
