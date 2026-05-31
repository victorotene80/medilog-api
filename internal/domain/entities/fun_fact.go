package entities

import "time"

type FunFact struct {
	ID                int64
	Title             *string
	Text              string
	Category          *string
	TargetCountryCode *string
	TargetAgeMin      *int
	TargetAgeMax      *int
	AllergyCategory   *int
	IsActive          bool
	CreatedAt         time.Time
}
