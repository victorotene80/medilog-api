package command

type AllergyCommand struct {
	Name        string
	Category    int
	Description *string
	//CreatedAt   time.Time
	//UpdatedAt time.Time
}

type UpdateAllergyCommand struct {
	ID          int64
	Name        string
	Category    int
	Description *string
}

type DeleteAllergyCommand struct {
	ID int64
}

type UserAllergyItemCommand struct {
	AllergyID   *int64
	Name        string
	Description *string
	Severity    *int16
	Category    int
}

type CreateUserAllergiesCommand struct {
	UserID    int64
	Allergies []UserAllergyItemCommand
}

type DeleteUserAllergyCommand struct {
	UserID   int64
	PublicID string
}
