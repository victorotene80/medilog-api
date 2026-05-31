package dto

import "time"

type UserProfileResponseDTO struct {
	ID                         string    `json:"id"`
	UserID                     string    `json:"user_id"`
	Weight                     *float64  `json:"weight,omitempty"`
	Height                     *float64  `json:"height,omitempty"`
	WeightUnit                 string    `json:"weight_unit"`
	TemperatureUnit            string    `json:"temperature_unit"`
	MedicationRemindersEnabled bool      `json:"medication_reminders_enabled"`
	RefillRemindersEnabled     bool      `json:"refill_reminders_enabled"`
	AppUpdatesEnabled          bool      `json:"app_updates_enabled"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
}
