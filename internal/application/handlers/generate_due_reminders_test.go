package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/readmodel"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"go.uber.org/zap"
)

func mustLoad(t *testing.T, name string) *time.Location {
	t.Helper()

	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("tzdata missing for %q: %v (the runtime image must ship tzdata, "+
			"or main.go must import _ \"time/tzdata\")", name, err)
	}

	return loc
}

func timePtr(t time.Time) *time.Time { return &t }

// The scheduler's whole job is turning "08:00 in the user's timezone" into an
// exact UTC instant. These are the cases where that goes wrong.
func TestMedicationSlots(t *testing.T) {
	tests := []struct {
		name      string
		timezone  string
		timeOfDay string
		frequency *string
		startDate *time.Time
		endDate   *time.Time
		from      time.Time
		to        time.Time
		want      []string // expected slots as RFC3339 in UTC
	}{
		{
			name:      "Lagos is UTC+1 year round",
			timezone:  "Africa/Lagos",
			timeOfDay: "08:00",
			from:      time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC),
			to:        time.Date(2026, 9, 8, 23, 59, 0, 0, time.UTC),
			want:      []string{"2026-09-08T07:00:00Z"},
		},
		{
			name:      "UTC user",
			timezone:  "UTC",
			timeOfDay: "08:00",
			from:      time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC),
			to:        time.Date(2026, 9, 8, 23, 59, 0, 0, time.UTC),
			want:      []string{"2026-09-08T08:00:00Z"},
		},
		{
			name:      "unknown timezone falls back to UTC rather than erroring",
			timezone:  "Mars/Olympus_Mons",
			timeOfDay: "08:00",
			from:      time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC),
			to:        time.Date(2026, 9, 8, 23, 59, 0, 0, time.UTC),
			want:      []string{"2026-09-08T08:00:00Z"},
		},
		{
			name:      "window spanning two local days yields two slots",
			timezone:  "UTC",
			timeOfDay: "08:00",
			from:      time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC),
			to:        time.Date(2026, 9, 9, 23, 59, 0, 0, time.UTC),
			want:      []string{"2026-09-08T08:00:00Z", "2026-09-09T08:00:00Z"},
		},
		{
			name:      "as_needed never generates a scheduled reminder",
			timezone:  "UTC",
			timeOfDay: "08:00",
			frequency: strPtr("as_needed"),
			from:      time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC),
			to:        time.Date(2026, 9, 10, 23, 59, 0, 0, time.UTC),
			want:      nil,
		},
		{
			name:      "once fires only on its start date",
			timezone:  "UTC",
			timeOfDay: "08:00",
			frequency: strPtr("once"),
			startDate: timePtr(time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)),
			from:      time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC),
			to:        time.Date(2026, 9, 11, 23, 59, 0, 0, time.UTC),
			want:      []string{"2026-09-09T08:00:00Z"},
		},
		{
			name:      "weekly fires only on the start date's weekday",
			timezone:  "UTC",
			timeOfDay: "08:00",
			frequency: strPtr("weekly"),
			// 2026-09-08 is a Tuesday.
			startDate: timePtr(time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)),
			from:      time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC),
			to:        time.Date(2026, 9, 16, 23, 59, 0, 0, time.UTC),
			want:      []string{"2026-09-08T08:00:00Z", "2026-09-15T08:00:00Z"},
		},
		{
			name:      "the course still runs on its final day",
			timezone:  "UTC",
			timeOfDay: "08:00",
			startDate: timePtr(time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)),
			endDate:   timePtr(time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)),
			from:      time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC),
			to:        time.Date(2026, 9, 11, 23, 59, 0, 0, time.UTC),
			want:      []string{"2026-09-08T08:00:00Z", "2026-09-09T08:00:00Z"},
		},
		{
			name:      "nothing before the start date",
			timezone:  "UTC",
			timeOfDay: "08:00",
			startDate: timePtr(time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)),
			from:      time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC),
			to:        time.Date(2026, 9, 9, 23, 59, 0, 0, time.UTC),
			want:      nil,
		},
		{
			name:      "New York across the spring-forward gap",
			timezone:  "America/New_York",
			timeOfDay: "08:00",
			// 2026-03-08 is the US DST transition; 08:00 local is well clear of
			// the 02:00 gap but the UTC offset changes from -05:00 to -04:00.
			from: time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC),
			to:   time.Date(2026, 3, 9, 23, 59, 0, 0, time.UTC),
			want: []string{
				"2026-03-07T13:00:00Z", // EST, UTC-5
				"2026-03-08T12:00:00Z", // EDT, UTC-4
				"2026-03-09T12:00:00Z",
			},
		},
		{
			name:      "London across the autumn fall-back",
			timezone:  "Europe/London",
			timeOfDay: "08:00",
			// 2026-10-25 is the UK transition back to GMT.
			from: time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC),
			to:   time.Date(2026, 10, 26, 23, 59, 0, 0, time.UTC),
			want: []string{
				"2026-10-24T07:00:00Z", // BST, UTC+1
				"2026-10-25T08:00:00Z", // GMT, UTC+0
				"2026-10-26T08:00:00Z",
			},
		},
		{
			name:      "malformed time of day yields nothing rather than panicking",
			timezone:  "UTC",
			timeOfDay: "not-a-time",
			from:      time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC),
			to:        time.Date(2026, 9, 8, 23, 59, 0, 0, time.UTC),
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := time.UTC
			if tt.timezone != "UTC" && tt.timezone != "Mars/Olympus_Mons" {
				loc = mustLoad(t, tt.timezone)
			}

			candidate := readmodel.DueMedicationCandidate{
				UserID:           1,
				Timezone:         tt.timezone,
				MedicationID:     10,
				MedicationTimeID: 20,
				MedicationName:   "Amoxicillin",
				Frequency:        tt.frequency,
				StartDate:        tt.startDate,
				EndDate:          tt.endDate,
				TimeOfDay:        tt.timeOfDay,
			}

			slots := medicationSlots(candidate, tt.from, tt.to, loc)

			got := make([]string, 0, len(slots))
			for _, s := range slots {
				got = append(got, s.UTC().Format(time.RFC3339))
			}

			if tt.want == nil {
				assert.Empty(t, got)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

// A local wall time inside a spring-forward gap does not exist. time.Date
// resolves it deterministically, but Go explicitly documents *which* side it
// lands on as unspecified — so this asserts the property that actually matters
// (exactly one slot, stable across runs) rather than pinning an instant that a
// future Go release could legitimately change.
func TestMedicationSlots_NonexistentLocalTimeYieldsExactlyOneStableSlot(t *testing.T) {
	newYork := mustLoad(t, "America/New_York")

	candidate := readmodel.DueMedicationCandidate{
		UserID:           1,
		Timezone:         "America/New_York",
		MedicationID:     10,
		MedicationTimeID: 20,
		MedicationName:   "Amoxicillin",
		TimeOfDay:        "02:30", // does not exist on 2026-03-08
	}

	from := time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 8, 23, 59, 0, 0, time.UTC)

	first := medicationSlots(candidate, from, to, newYork)
	assert.Len(t, first, 1, "the skipped hour must still produce exactly one reminder")

	// Stability is what keeps the dedupe key from flapping between ticks and
	// emitting a second notification for the same dose.
	second := medicationSlots(candidate, from, to, newYork)
	assert.Equal(t, first, second, "resolution must be deterministic across calls")
}

// The dedupe key must be scoped to the user's *local* day. Keying off the UTC
// date would split one local day in two for users far from UTC, producing two
// reminders for a single dose.
func TestBuildMedicationNotification_DedupeKeyUsesLocalDate(t *testing.T) {
	lagos := mustLoad(t, "Africa/Lagos")

	candidate := readmodel.DueMedicationCandidate{
		UserID:           7,
		Timezone:         "Africa/Lagos",
		MedicationID:     42,
		MedicationTimeID: 99,
		MedicationName:   "Amoxicillin",
		Dosage:           strPtr("500mg"),
		TimeOfDay:        "00:30",
	}

	// 00:30 on the 9th in Lagos is 23:30 on the 8th in UTC.
	slot := time.Date(2026, 9, 9, 0, 30, 0, 0, lagos).UTC()
	assert.Equal(t, "2026-09-08T23:30:00Z", slot.Format(time.RFC3339))

	n := buildMedicationNotification(candidate, slot, lagos)

	assert.Equal(t, "med:42:99:2026-09-09", *n.DedupeKey,
		"the key must use the local date, not the UTC date")
	assert.Equal(t, int64(7), n.UserID)
	assert.Equal(t, "Time for Amoxicillin", n.Title)
	assert.Equal(t, "500mg · 00:30", n.Body)
	assert.Equal(t, notificationTypeMedication, n.Type)
	assert.Equal(t, notificationStatusUnread, n.Status)
	assert.Equal(t, "2026-09-09", n.Metadata["local_date"])
}

func TestBuildMedicationNotification_OmitsMissingDosage(t *testing.T) {
	candidate := readmodel.DueMedicationCandidate{
		UserID:         7,
		MedicationName: "Amoxicillin",
		TimeOfDay:      "08:00",
	}

	n := buildMedicationNotification(
		candidate, time.Date(2026, 9, 8, 8, 0, 0, 0, time.UTC), time.UTC)

	assert.Equal(t, "08:00", n.Body, "no dosage means no leading separator")
}

func TestBuildAppointmentNotification(t *testing.T) {
	visitDate := time.Date(2026, 9, 9, 14, 0, 0, 0, time.UTC)

	candidate := readmodel.DueAppointmentCandidate{
		UserID:       3,
		Timezone:     "UTC",
		VisitID:      55,
		HospitalName: strPtr("Lagos General"),
		Doctor:       strPtr("Dr Ade"),
		VisitDate:    visitDate,
	}

	n := buildAppointmentNotification(candidate, "2026-09-09")

	assert.Equal(t, "visit:55:2026-09-09:24h", *n.DedupeKey)
	assert.Equal(t, "Appointment at Lagos General", n.Title)
	assert.Equal(t, "Tomorrow with Dr Ade", n.Body)
	assert.Equal(t, notificationTypeAppointment, n.Type)
	// Fires one lead-time before the visit.
	assert.Equal(t, visitDate.Add(-appointmentLeadTime), *n.ScheduledAt)
}

func TestLocationCache_FallsBackAndCaches(t *testing.T) {
	cache := newLocationCache(nil)

	assert.Equal(t, time.UTC, cache.get(""), "empty timezone means UTC")
	assert.Equal(t, time.UTC, cache.get("Nope/Nowhere"), "a bad zone must not error the run")

	lagos := cache.get("Africa/Lagos")
	assert.Equal(t, "Africa/Lagos", lagos.String())
	assert.Same(t, lagos, cache.get("Africa/Lagos"), "second lookup is served from cache")
}

// --- pagination ---------------------------------------------------------
//
// The scan loop had no coverage, which is where two silent data-loss bugs
// lived: the medication cursor advanced on medication id alone even though a
// row is a (medication, medication_time) pair, and the appointment scan took a
// single page and never looped.

type fakeReminderRepo struct {
	medPages    [][]readmodel.DueMedicationCandidate
	medCursors  [][2]int64
	apptPages   [][]readmodel.DueAppointmentCandidate
	apptCursors []int64
}

func (f *fakeReminderRepo) FindMedicationCandidates(
	_ context.Context, _, _ time.Time, _ int, afterMedID, afterTimeID int64,
) ([]readmodel.DueMedicationCandidate, error) {
	f.medCursors = append(f.medCursors, [2]int64{afterMedID, afterTimeID})
	if len(f.medPages) == 0 {
		return nil, nil
	}
	page := f.medPages[0]
	f.medPages = f.medPages[1:]
	return page, nil
}

func (f *fakeReminderRepo) FindAppointmentCandidates(
	_ context.Context, _, _ time.Time, _ int, _ time.Time, afterVisitID int64,
) ([]readmodel.DueAppointmentCandidate, error) {
	f.apptCursors = append(f.apptCursors, afterVisitID)
	if len(f.apptPages) == 0 {
		return nil, nil
	}
	page := f.apptPages[0]
	f.apptPages = f.apptPages[1:]
	return page, nil
}

func (f *fakeReminderRepo) TryAcquireScanLock(context.Context, int64) (bool, func(), error) {
	return true, func() {}, nil
}

type countingNotificationRepo struct {
	repository.NotificationRepository
	saved []string
}

func (c *countingNotificationRepo) SaveIfNotExists(_ context.Context, n *entities.Notification) (bool, error) {
	c.saved = append(c.saved, derefString(n.DedupeKey))
	return true, nil
}

// A page boundary falling between two dose times of one medication must not
// lose the remaining doses. Cursoring on medication id alone did exactly that.
func TestGenerateMedicationReminders_PageBoundaryInsideMedication(t *testing.T) {
	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)

	row := func(medID, timeID int64, at string) readmodel.DueMedicationCandidate {
		return readmodel.DueMedicationCandidate{
			UserID: 1, Timezone: "UTC",
			MedicationID: medID, MedicationTimeID: timeID,
			MedicationName: "Metformin", TimeOfDay: at,
		}
	}

	repo := &fakeReminderRepo{medPages: [][]readmodel.DueMedicationCandidate{
		// limit is 2, so this page is "full" and the loop must ask for more —
		// and medication 42 still has an unseen 14:00 dose.
		{row(42, 100, "08:00"), row(42, 101, "12:00")},
		{row(42, 102, "14:00")},
	}}
	notifs := &countingNotificationRepo{}

	h := NewGenerateDueRemindersHandler(repo, notifs, nil, func() time.Time { return from })
	result := &dto.ReminderRunResultDTO{}

	err := h.generateMedicationReminders(
		context.Background(),
		command.GenerateDueRemindersCommand{From: from, To: to, Limit: 2},
		newLocationCache(zap.NewNop()),
		result,
	)

	require.NoError(t, err)
	assert.Len(t, repo.medCursors, 2, "a full page must be followed by another fetch")
	assert.Equal(t, [2]int64{0, 0}, repo.medCursors[0])
	assert.Equal(t, [2]int64{42, 101}, repo.medCursors[1],
		"cursor must carry the medication_time id, or medication 42's later doses are skipped")
	assert.Equal(t, 3, result.Scanned)
	assert.Len(t, notifs.saved, 3, "all three dose times must produce a reminder")
}

// The appointment scan previously took one page and stopped, so every visit
// past the batch size was never reminded — on any tick.
func TestGenerateAppointmentReminders_PaginatesBeyondFirstPage(t *testing.T) {
	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	visit := func(id int64) readmodel.DueAppointmentCandidate {
		return readmodel.DueAppointmentCandidate{
			UserID: 1, Timezone: "UTC", VisitID: id,
			VisitDate: from.Add(time.Duration(id) * time.Hour),
		}
	}

	repo := &fakeReminderRepo{apptPages: [][]readmodel.DueAppointmentCandidate{
		{visit(1), visit(2)},
		{visit(3)},
	}}
	notifs := &countingNotificationRepo{}

	h := NewGenerateDueRemindersHandler(repo, notifs, nil, func() time.Time { return from })
	result := &dto.ReminderRunResultDTO{}

	err := h.generateAppointmentReminders(
		context.Background(),
		command.GenerateDueRemindersCommand{From: from, To: from.Add(24 * time.Hour), Limit: 2},
		newLocationCache(zap.NewNop()),
		result,
	)

	require.NoError(t, err)
	assert.Len(t, repo.apptCursors, 2, "a full page must be followed by another fetch")
	assert.Equal(t, int64(0), repo.apptCursors[0])
	assert.Equal(t, int64(2), repo.apptCursors[1])
	assert.Equal(t, 3, result.Scanned)
	assert.Len(t, notifs.saved, 3)
}
