package command

// UpdateNotificationPreferencesCommand is a partial update of the caller's
// delivery preferences. Every field is a pointer: nil means "leave unchanged",
// so a client can save a single toggle without echoing the whole set back.
type UpdateNotificationPreferencesCommand struct {
	UserID int64

	MedicationRemindersEnabled  *bool
	RefillRemindersEnabled      *bool
	AppointmentRemindersEnabled *bool
	AIHealthTipsEnabled         *bool
	SupportUpdatesEnabled       *bool
	AppUpdatesEnabled           *bool

	PushEnabled     *bool
	EmailEnabled    *bool
	SMSEnabled      *bool
	WhatsAppEnabled *bool

	// Timezone is an IANA name; validated with time.LoadLocation.
	Timezone *string
}
