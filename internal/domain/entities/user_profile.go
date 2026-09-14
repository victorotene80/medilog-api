package entities

import "time"

const (
	DefaultWeightUnit      = "kg"
	DefaultTemperatureUnit = "celsius"
	DefaultTimezone        = "UTC"
	// DefaultAIQuestionsTotal is 10 to agree with the column default in
	// 000001_init.up.sql and the gorm tag on models.UserProfileModel. It used to
	// be 3, so users created through the API got 3 while rows created by any
	// other path got 10. Resolving toward 10 needs no backfill and never shrinks
	// an existing user's allowance.
	DefaultAIQuestionsTotal = 10
)

type UserProfile struct {
	ID              int64
	UserID          int64
	Weight          *float64
	Height          *float64
	WeightUnit      string
	TemperatureUnit string
	// Timezone is an IANA name (e.g. "Africa/Lagos") used to place reminders at
	// the user's local wall-clock time. Never derived from country code.
	Timezone                    string
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
		Timezone:                    DefaultTimezone,
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

// AIQuestionsRemaining is what the client renders in the quota UI. Pro accounts
// are uncapped, reported as 0 remaining only when the caller checks AIIsPro
// first; callers should treat AIIsPro as "unlimited".
func (p *UserProfile) AIQuestionsRemaining() int {
	remaining := p.AIQuestionsTotal - p.AIQuestionsUsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// EnsureDailyWindow resets the quota when the stored reset stamp predates the
// current UTC day, and reports whether it changed anything so the caller can
// persist. This is a lazy reset performed on the read and send paths, which is
// what lets a daily window work with no scheduler.
//
// The boundary is UTC midnight rather than the user's local midnight: the quota
// is an abuse control, and a per-user-timezone window lets someone crossing
// zones collect two resets inside 24 hours.
func (p *UserProfile) EnsureDailyWindow(now time.Time) bool {
	utcNow := now.UTC()
	startOfDay := time.Date(utcNow.Year(), utcNow.Month(), utcNow.Day(), 0, 0, 0, 0, time.UTC)

	if p.AIQuestionsResetAt != nil && !p.AIQuestionsResetAt.UTC().Before(startOfDay) {
		return false
	}

	p.ResetAIQuota(startOfDay)

	return true
}

// AIQuotaResetsAt is the next UTC midnight — when EnsureDailyWindow will next
// clear the counter.
func (p *UserProfile) AIQuotaResetsAt(now time.Time) time.Time {
	utcNow := now.UTC()
	return time.Date(utcNow.Year(), utcNow.Month(), utcNow.Day(), 0, 0, 0, 0, time.UTC).
		AddDate(0, 0, 1)
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
