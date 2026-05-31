package command

type CreateFunFactCommand struct {
	Title             *string
	Text              string
	Category          *string
	TargetCountryCode *string
	TargetAgeMin      *int
	TargetAgeMax      *int
	AllergyCategory   *int
	IsActive          *bool
}

type UpdateFunFactCommand struct {
	ID                int64
	Title             *string
	Text              string
	Category          *string
	TargetCountryCode *string
	TargetAgeMin      *int
	TargetAgeMax      *int
	AllergyCategory   *int
	IsActive          *bool
}

type DeleteFunFactCommand struct {
	ID int64
}
