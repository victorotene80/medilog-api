package command

type RegisterCommand struct {
	FirstName   string
	LastName    string
	Email       *string
	Phone       *string
	Password    string
	DOB         string
	Sex         int
	BloodType   string
	CountryCode string
}
