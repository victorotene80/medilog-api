package response

// NotificationPreferencesResponse is the settings screen's payload.
//
// These flags are honoured by the reminder scheduler. refill_reminders_enabled
// is stored and respected but currently generates nothing, because medications
// carry no quantity or days-supply column to compute a refill date from.
type NotificationPreferencesResponse struct {
	MedicationRemindersEnabled  bool   `json:"medication_reminders_enabled"`
	RefillRemindersEnabled      bool   `json:"refill_reminders_enabled"`
	AppointmentRemindersEnabled bool   `json:"appointment_reminders_enabled"`
	AIHealthTipsEnabled         bool   `json:"ai_health_tips_enabled"`
	SupportUpdatesEnabled       bool   `json:"support_updates_enabled"`
	AppUpdatesEnabled           bool   `json:"app_updates_enabled"`
	PushEnabled                 bool   `json:"push_enabled"`
	EmailEnabled                bool   `json:"email_enabled"`
	SMSEnabled                  bool   `json:"sms_enabled"`
	WhatsAppEnabled             bool   `json:"whatsapp_enabled"`
	Timezone                    string `json:"timezone"`
}
