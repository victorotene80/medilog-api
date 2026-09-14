package command

type EmergencyContactCommand struct {
	UserID       int64
	Name         string
	Relationship string
	Phone        string
	CountryCode  string
	IsPrimary    bool
}

// UpdateEmergencyContactCommand fully replaces a contact addressed by its
// public id.
type UpdateEmergencyContactCommand struct {
	UserID       int64
	PublicID     string
	Name         string
	Relationship string
	Phone        string
	CountryCode  string
	IsPrimary    bool
}

type DeleteEmergencyContactCommand struct {
	UserID   int64
	PublicID string
}
