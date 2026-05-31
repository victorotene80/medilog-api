package dto

import "time"

type NotificationResponseDTO struct {
	ID                string     `json:"id"`
	UserID            string     `json:"user_id"`
	Title             string     `json:"title"`
	Body              string     `json:"body"`
	Type              string     `json:"type"`
	Channel           string     `json:"channel"`
	Status            string     `json:"status"`
	RelatedEntityType *string    `json:"related_entity_type,omitempty"`
	RelatedEntityID   *string    `json:"related_entity_id,omitempty"`
	ActionURL         *string    `json:"action_url,omitempty"`
	ImageURL          *string    `json:"image_url,omitempty"`
	ScheduledAt       *time.Time `json:"scheduled_at,omitempty"`
	SentAt            *time.Time `json:"sent_at,omitempty"`
	ReadAt            *time.Time `json:"read_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type ListNotificationsResponseDTO struct {
	Notifications []*NotificationResponseDTO `json:"notifications"`
	NextCursor    *string                    `json:"next_cursor,omitempty"`
}

type NotificationPreferenceResponseDTO struct {
	ID                          string     `json:"id"`
	UserID                      string     `json:"user_id"`
	MedicationRemindersEnabled  bool       `json:"medication_reminders_enabled"`
	RefillRemindersEnabled      bool       `json:"refill_reminders_enabled"`
	AppointmentRemindersEnabled bool       `json:"appointment_reminders_enabled"`
	AIHealthTipsEnabled         bool       `json:"ai_health_tips_enabled"`
	SupportUpdatesEnabled       bool       `json:"support_updates_enabled"`
	AppUpdatesEnabled           bool       `json:"app_updates_enabled"`
	PushEnabled                 bool       `json:"push_enabled"`
	EmailEnabled                bool       `json:"email_enabled"`
	WhatsappEnabled             bool       `json:"whatsapp_enabled"`
	SMSEnabled                  bool       `json:"sms_enabled"`
	QuietHoursEnabled           bool       `json:"quiet_hours_enabled"`
	QuietHoursStart             *time.Time `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd               *time.Time `json:"quiet_hours_end,omitempty"`
	CreatedAt                   time.Time  `json:"created_at"`
	UpdatedAt                   time.Time  `json:"updated_at"`
}
