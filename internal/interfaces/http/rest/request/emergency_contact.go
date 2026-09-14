package request

type CreateEmergencyContactRequest struct {
	Name         string `json:"name"         validate:"required,min=1,max=200"`
	Relationship string `json:"relationship" validate:"required,min=1,max=100"`
	Phone        string `json:"phone"        validate:"required,e164"`
	CountryCode  string `json:"country_code" validate:"required,min=1,max=4"`
	//Address      string `json:"address"      validate:"omitempty,max=500"`
	IsPrimary bool `json:"is_primary"`
}

// UpdateEmergencyContactRequest fully replaces a contact. Every field is
// required: this is a PUT, not a partial update.
type UpdateEmergencyContactRequest struct {
	Name         string `json:"name"         validate:"required,min=1,max=200"`
	Relationship string `json:"relationship" validate:"required,min=1,max=100"`
	Phone        string `json:"phone"        validate:"required,e164"`
	CountryCode  string `json:"country_code" validate:"required,min=1,max=4"`
	IsPrimary    bool   `json:"is_primary"`
}
