package command

type EmergencyContactCommand struct {
	UserID       int64
	Name         string
	Relationship string
	Phone        string
	CountryCode  string
}
