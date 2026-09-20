package request

import "time"

// UpdateUserRequest is a partial update of the caller's own profile.
//
// Email and phone are intentionally absent. decodeJSON sets
// DisallowUnknownFields, so a body containing "email" or "phone" is rejected
// with 400 rather than silently ignored — changing either has to go through the
// existing OTP-verified flow.
//
// Note a nil pointer means "leave unchanged"; there is no way to clear a
// nullable field through this endpoint.
type UpdateUserRequest struct {
	FirstName *string    `json:"first_name"       validate:"omitempty,min=1,max=100"`
	LastName  *string    `json:"last_name"        validate:"omitempty,min=1,max=100"`
	DOB       *time.Time `json:"dob"              validate:"omitempty"`
	// Sex: 1=male 2=female 3=other
	Sex         *int    `json:"sex"          validate:"omitempty,min=1,max=3"`
	BloodType   *string `json:"blood_type"   validate:"omitempty,oneof=A+ A- B+ B- O+ O- AB+ AB-"`
	AvatarURL   *string `json:"avatar_url"   validate:"omitempty,url,max=2048"`
	CountryCode *string `json:"country_code" validate:"omitempty,len=2,alpha"`

	Height          *float64 `json:"height"           validate:"omitempty,gt=0,lte=300"`
	Weight          *float64 `json:"weight"           validate:"omitempty,gt=0,lte=700"`
	WeightUnit      *string  `json:"weight_unit"      validate:"omitempty,oneof=kg lb"`
	TemperatureUnit *string  `json:"temperature_unit" validate:"omitempty,oneof=celsius fahrenheit"`

	// Timezone is an IANA name such as "Africa/Lagos". Also settable on
	// PATCH /users/me/notification-preferences; it is accepted here too because
	// this is the endpoint a client reaches for when filling in a profile, and
	// leaving it out meant new users silently kept the UTC default and received
	// every reminder at the wrong local time.
	Timezone *string `json:"timezone" validate:"omitempty,min=1,max=64"`
}

// UpdateNotificationPreferencesRequest is a partial update; nil means
// "leave unchanged" so the client can save one toggle at a time.
type UpdateNotificationPreferencesRequest struct {
	MedicationRemindersEnabled  *bool `json:"medication_reminders_enabled"`
	RefillRemindersEnabled      *bool `json:"refill_reminders_enabled"`
	AppointmentRemindersEnabled *bool `json:"appointment_reminders_enabled"`
	AIHealthTipsEnabled         *bool `json:"ai_health_tips_enabled"`
	SupportUpdatesEnabled       *bool `json:"support_updates_enabled"`
	AppUpdatesEnabled           *bool `json:"app_updates_enabled"`

	PushEnabled     *bool `json:"push_enabled"`
	EmailEnabled    *bool `json:"email_enabled"`
	SMSEnabled      *bool `json:"sms_enabled"`
	WhatsAppEnabled *bool `json:"whatsapp_enabled"`

	// Timezone is an IANA name such as "Africa/Lagos". The reminder scheduler
	// resolves it on every tick, so it is validated on write.
	Timezone *string `json:"timezone" validate:"omitempty,min=1,max=64"`
}
