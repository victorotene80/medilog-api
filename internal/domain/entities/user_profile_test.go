package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserProfile_HasAIQuotaRemaining(t *testing.T) {
	tests := []struct {
		name  string
		used  int
		total int
		isPro bool
		want  bool
	}{
		{name: "under the cap", used: 3, total: 10, want: true},
		{name: "one left", used: 9, total: 10, want: true},
		{name: "exactly at the cap", used: 10, total: 10, want: false},
		{name: "over the cap", used: 11, total: 10, want: false},
		{name: "pro is uncapped even past the cap", used: 99, total: 10, isPro: true, want: true},
		{name: "zero allowance", used: 0, total: 0, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &UserProfile{
				AIQuestionsUsed:  tt.used,
				AIQuestionsTotal: tt.total,
				AIIsPro:          tt.isPro,
			}

			assert.Equal(t, tt.want, p.HasAIQuotaRemaining())
		})
	}
}

func TestUserProfile_EnsureDailyWindow(t *testing.T) {
	now := time.Date(2026, 9, 8, 14, 30, 0, 0, time.UTC)
	startOfToday := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)

	ptr := func(tm time.Time) *time.Time { return &tm }

	tests := []struct {
		name        string
		resetAt     *time.Time
		usedBefore  int
		wantChanged bool
		wantUsed    int
	}{
		{
			name:        "never reset before",
			resetAt:     nil,
			usedBefore:  4,
			wantChanged: true,
			wantUsed:    0,
		},
		{
			name:        "already reset today",
			resetAt:     ptr(startOfToday),
			usedBefore:  4,
			wantChanged: false,
			wantUsed:    4,
		},
		{
			name:        "reset later today still counts as today",
			resetAt:     ptr(now.Add(-time.Hour)),
			usedBefore:  7,
			wantChanged: false,
			wantUsed:    7,
		},
		{
			name:        "reset yesterday",
			resetAt:     ptr(startOfToday.AddDate(0, 0, -1)),
			usedBefore:  10,
			wantChanged: true,
			wantUsed:    0,
		},
		{
			name:        "one second before midnight is the previous window",
			resetAt:     ptr(startOfToday.Add(-time.Second)),
			usedBefore:  10,
			wantChanged: true,
			wantUsed:    0,
		},
		{
			name:        "reset long ago",
			resetAt:     ptr(startOfToday.AddDate(0, -3, 0)),
			usedBefore:  10,
			wantChanged: true,
			wantUsed:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &UserProfile{
				AIQuestionsUsed:    tt.usedBefore,
				AIQuestionsTotal:   10,
				AIQuestionsResetAt: tt.resetAt,
			}

			changed := p.EnsureDailyWindow(now)

			assert.Equal(t, tt.wantChanged, changed)
			assert.Equal(t, tt.wantUsed, p.AIQuestionsUsed)

			if tt.wantChanged {
				// The stamp is pinned to the start of the UTC day so the next
				// day's comparison is unambiguous.
				assert.Equal(t, startOfToday, p.AIQuestionsResetAt.UTC())
			}
		})
	}
}

// A non-UTC clock must not shift which day the window belongs to.
func TestUserProfile_EnsureDailyWindow_NormalizesNonUTCClock(t *testing.T) {
	lagos, err := time.LoadLocation("Africa/Lagos")
	assert.NoError(t, err)

	// 00:30 on the 9th in Lagos (UTC+1) is still 23:30 on the 8th in UTC.
	localNow := time.Date(2026, 9, 9, 0, 30, 0, 0, lagos)
	startOfUTCDay := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)

	p := &UserProfile{
		AIQuestionsUsed:    5,
		AIQuestionsTotal:   10,
		AIQuestionsResetAt: &startOfUTCDay,
	}

	assert.False(t, p.EnsureDailyWindow(localNow), "still inside the same UTC day")
	assert.Equal(t, 5, p.AIQuestionsUsed)
}

func TestUserProfile_AIQuotaResetsAt(t *testing.T) {
	now := time.Date(2026, 9, 8, 14, 30, 0, 0, time.UTC)

	assert.Equal(t,
		time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC),
		(&UserProfile{}).AIQuotaResetsAt(now),
	)
}

func TestUserProfile_AIQuestionsRemaining(t *testing.T) {
	assert.Equal(t, 7, (&UserProfile{AIQuestionsUsed: 3, AIQuestionsTotal: 10}).AIQuestionsRemaining())
	assert.Equal(t, 0, (&UserProfile{AIQuestionsUsed: 10, AIQuestionsTotal: 10}).AIQuestionsRemaining())
	assert.Equal(t, 0, (&UserProfile{AIQuestionsUsed: 12, AIQuestionsTotal: 10}).AIQuestionsRemaining(),
		"never reports a negative remaining count")
}

// The Go default must agree with the SQL column default and the gorm tag,
// otherwise a user's allowance depends on which code path created their row.
func TestDefaultAIQuestionsTotal_MatchesSchemaDefault(t *testing.T) {
	assert.Equal(t, 10, DefaultAIQuestionsTotal)
	assert.Equal(t, 10, NewDefaultUserProfile(1).AIQuestionsTotal)
}

func TestNewDefaultUserProfile_DefaultsToUTC(t *testing.T) {
	assert.Equal(t, "UTC", NewDefaultUserProfile(1).Timezone)
}
