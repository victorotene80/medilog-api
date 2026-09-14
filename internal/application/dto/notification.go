package dto

import "time"

type NotificationResponseDTO struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	Title       string         `json:"title"`
	Body        string         `json:"body"`
	Type        string         `json:"type"`
	Channel     *string        `json:"channel,omitempty"`
	Status      string         `json:"status"`
	ImageURL    *string        `json:"image_url,omitempty"`
	ScheduledAt *time.Time     `json:"scheduled_at,omitempty"`
	SentAt      *time.Time     `json:"sent_at,omitempty"`
	ReadAt      *time.Time     `json:"read_at,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type ListNotificationsResponseDTO struct {
	Notifications []*NotificationResponseDTO `json:"notifications"`
	NextCursor    *string                    `json:"next_cursor,omitempty"`
}

// NotificationPreferencesDTO is the settings screen's payload.
//
// The previous NotificationPreferenceResponseDTO carried quiet-hours fields
// that had no backing columns and no callers. With no push transport there is
// nothing for quiet hours to suppress — the client polls an inbox and renders
// it whenever it likes — so shipping them would have been an API that lies.
type NotificationPreferencesDTO struct {
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
