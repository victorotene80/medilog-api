package entities

import "time"

const (
	DefaultWeightUnit       = "kg"
	DefaultTemperatureUnit  = "celsius"
	DefaultAIQuestionsTotal = 3
)

type UserProfile struct {
	ID                          int64
	UserID                      int64
	Weight                      *float64
	Height                      *float64
	WeightUnit                  string
	TemperatureUnit             string
	AIQuestionsUsed             int
	AIQuestionsTotal            int
	AIIsPro                     bool
	AIQuestionsResetAt          *time.Time
	MedicationRemindersEnabled  bool
	RefillRemindersEnabled      bool
	AppointmentRemindersEnabled bool
	AIHealthTipsEnabled         bool
	SupportUpdatesEnabled       bool
	AppUpdatesEnabled           bool
	PushEnabled                 bool
	EmailEnabled                bool
	SMSEnabled                  bool
	WhatsAppEnabled             bool
}

func NewDefaultUserProfile(userID int64) *UserProfile {
	return &UserProfile{
		UserID:                      userID,
		Weight:                      nil,
		Height:                      nil,
		WeightUnit:                  DefaultWeightUnit,
		TemperatureUnit:             DefaultTemperatureUnit,
		AIQuestionsUsed:             0,
		AIQuestionsTotal:            DefaultAIQuestionsTotal,
		AIIsPro:                     false,
		AIQuestionsResetAt:          nil,
		MedicationRemindersEnabled:  true,
		RefillRemindersEnabled:      true,
		AppointmentRemindersEnabled: true,
		AIHealthTipsEnabled:         true,
		SupportUpdatesEnabled:       true,
		AppUpdatesEnabled:           true,
		PushEnabled:                 true,
		EmailEnabled:                true,
		SMSEnabled:                  false,
		WhatsAppEnabled:             false,
	}
}

func (p *UserProfile) HasAIQuotaRemaining() bool {
	return p.AIIsPro || p.AIQuestionsUsed < p.AIQuestionsTotal
}

func (p *UserProfile) ConsumeAIQuestion() bool {
	if !p.HasAIQuotaRemaining() {
		return false
	}
	if !p.AIIsPro {
		p.AIQuestionsUsed++
	}
	return true
}

func (p *UserProfile) ResetAIQuota(now time.Time) {
	p.AIQuestionsUsed = 0
	p.AIQuestionsResetAt = &now
}
