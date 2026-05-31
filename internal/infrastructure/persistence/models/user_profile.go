package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type UserProfileModel struct {
	ID                          int64      `gorm:"column:id;primaryKey;autoIncrement"`
	UserID                      int64      `gorm:"column:user_id;not null;uniqueIndex"`
	Weight                      *float64   `gorm:"column:weight"`
	Height                      *float64   `gorm:"column:height"`
	WeightUnit                  string     `gorm:"column:weight_unit;default:kg"`
	TemperatureUnit             string     `gorm:"column:temperature_unit;default:celsius"`
	AIQuestionsUsed             int        `gorm:"column:ai_questions_used;not null;default:0"`
	AIQuestionsTotal            int        `gorm:"column:ai_questions_total;not null;default:10"`
	AIIsPro                     bool       `gorm:"column:ai_is_pro;not null;default:false"`
	AIQuestionsResetAt          *time.Time `gorm:"column:ai_questions_reset_at"`
	MedicationRemindersEnabled  bool       `gorm:"column:medication_reminders_enabled;not null;default:true"`
	RefillRemindersEnabled      bool       `gorm:"column:refill_reminders_enabled;not null;default:true"`
	AppointmentRemindersEnabled bool       `gorm:"column:appointment_reminders_enabled;not null;default:true"`
	AIHealthTipsEnabled         bool       `gorm:"column:ai_health_tips_enabled;not null;default:true"`
	SupportUpdatesEnabled       bool       `gorm:"column:support_updates_enabled;not null;default:true"`
	AppUpdatesEnabled           bool       `gorm:"column:app_updates_enabled;not null;default:true"`
	PushEnabled                 bool       `gorm:"column:push_enabled;not null;default:true"`
	EmailEnabled                bool       `gorm:"column:email_enabled;not null;default:true"`
	SMSEnabled                  bool       `gorm:"column:sms_enabled;not null;default:false"`
	WhatsAppEnabled             bool       `gorm:"column:whatsapp_enabled;not null;default:false"`
}

func (UserProfileModel) TableName() string {
	return "user_profiles"
}

func UserProfileModelToEntity(m *UserProfileModel) *entities.UserProfile {
	if m == nil {
		return nil
	}

	return &entities.UserProfile{
		ID:                          m.ID,
		UserID:                      m.UserID,
		Weight:                      m.Weight,
		Height:                      m.Height,
		WeightUnit:                  m.WeightUnit,
		TemperatureUnit:             m.TemperatureUnit,
		AIQuestionsUsed:             m.AIQuestionsUsed,
		AIQuestionsTotal:            m.AIQuestionsTotal,
		AIIsPro:                     m.AIIsPro,
		AIQuestionsResetAt:          m.AIQuestionsResetAt,
		MedicationRemindersEnabled:  m.MedicationRemindersEnabled,
		RefillRemindersEnabled:      m.RefillRemindersEnabled,
		AppointmentRemindersEnabled: m.AppointmentRemindersEnabled,
		AIHealthTipsEnabled:         m.AIHealthTipsEnabled,
		SupportUpdatesEnabled:       m.SupportUpdatesEnabled,
		AppUpdatesEnabled:           m.AppUpdatesEnabled,
		PushEnabled:                 m.PushEnabled,
		EmailEnabled:                m.EmailEnabled,
		SMSEnabled:                  m.SMSEnabled,
		WhatsAppEnabled:             m.WhatsAppEnabled,
	}
}

func UserProfileEntityToModel(e *entities.UserProfile) *UserProfileModel {
	if e == nil {
		return nil
	}

	return &UserProfileModel{
		ID:                          e.ID,
		UserID:                      e.UserID,
		Weight:                      e.Weight,
		Height:                      e.Height,
		WeightUnit:                  e.WeightUnit,
		TemperatureUnit:             e.TemperatureUnit,
		AIQuestionsUsed:             e.AIQuestionsUsed,
		AIQuestionsTotal:            e.AIQuestionsTotal,
		AIIsPro:                     e.AIIsPro,
		AIQuestionsResetAt:          e.AIQuestionsResetAt,
		MedicationRemindersEnabled:  e.MedicationRemindersEnabled,
		RefillRemindersEnabled:      e.RefillRemindersEnabled,
		AppointmentRemindersEnabled: e.AppointmentRemindersEnabled,
		AIHealthTipsEnabled:         e.AIHealthTipsEnabled,
		SupportUpdatesEnabled:       e.SupportUpdatesEnabled,
		AppUpdatesEnabled:           e.AppUpdatesEnabled,
		PushEnabled:                 e.PushEnabled,
		EmailEnabled:                e.EmailEnabled,
		SMSEnabled:                  e.SMSEnabled,
		WhatsAppEnabled:             e.WhatsAppEnabled,
	}
}
