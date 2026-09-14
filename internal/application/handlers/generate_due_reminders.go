package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/readmodel"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"
	"go.uber.org/zap"
)

// reminderScanLockKey is an arbitrary but fixed advisory-lock id. Only this
// scheduler uses it.
const reminderScanLockKey int64 = 4823001

// appointmentLeadTime is how far ahead of a visit its reminder fires.
const appointmentLeadTime = 24 * time.Hour

const (
	notificationTypeMedication  = "medication_reminder"
	notificationTypeAppointment = "appointment_reminder"
	notificationChannelInApp    = "in_app"
	notificationStatusUnread    = "unread"
)

// GenerateDueRemindersHandler turns medication schedules and upcoming visits
// into notification rows.
//
// Expansion from "this medication is taken at 08:00" to "this exact UTC
// instant" happens here rather than in SQL, because the interesting cases —
// DST boundaries, per-user timezones, weekly/once frequencies — are far easier
// to test as a table of Go cases than as a query.
type GenerateDueRemindersHandler struct {
	reminders     repository.ReminderRepository
	notifications repository.NotificationRepository
	logger        *zap.Logger
	clock         func() time.Time
}

func NewGenerateDueRemindersHandler(
	reminders repository.ReminderRepository,
	notifications repository.NotificationRepository,
	logger *zap.Logger,
	clock func() time.Time,
) *GenerateDueRemindersHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &GenerateDueRemindersHandler{
		reminders:     reminders,
		notifications: notifications,
		logger:        logger,
		clock:         clock,
	}
}

func (h *GenerateDueRemindersHandler) Handle(
	ctx context.Context,
	cmd command.GenerateDueRemindersCommand,
) (*dto.ReminderRunResultDTO, error) {
	result := &dto.ReminderRunResultDTO{}

	if !cmd.To.After(cmd.From) {
		return result, nil
	}

	// Several replicas running this at once is safe but wasteful; the advisory
	// lock keeps all but one from scanning. Correctness comes from the dedupe
	// index, so failing to take the lock is not an error.
	acquired, release, err := h.reminders.TryAcquireScanLock(ctx, reminderScanLockKey)
	if err != nil {
		return nil, fmt.Errorf("acquire reminder scan lock: %w", err)
	}
	if !acquired {
		result.LockNotAcquired = true
		return result, nil
	}
	defer release()

	// time.LoadLocation touches the filesystem, and the same handful of zones
	// recur across thousands of rows.
	locations := newLocationCache(h.logger)

	if err := h.generateMedicationReminders(ctx, cmd, locations, result); err != nil {
		return nil, err
	}

	if err := h.generateAppointmentReminders(ctx, cmd, locations, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (h *GenerateDueRemindersHandler) generateMedicationReminders(
	ctx context.Context,
	cmd command.GenerateDueRemindersCommand,
	locations *locationCache,
	result *dto.ReminderRunResultDTO,
) error {
	limit := cmd.Limit
	if limit <= 0 {
		limit = 500
	}

	// The cursor is the (medication, medication_time) pair the last row carried,
	// advanced from the final row rather than from a max over the page: a page
	// can end mid-medication, and cursoring on medication id alone would skip
	// that medication's remaining dose times on every tick.
	var afterMedicationID, afterMedicationTimeID int64

	for {
		candidates, err := h.reminders.FindMedicationCandidates(
			ctx, cmd.From, cmd.To, limit, afterMedicationID, afterMedicationTimeID,
		)
		if err != nil {
			return fmt.Errorf("find medication candidates: %w", err)
		}

		if len(candidates) == 0 {
			return nil
		}

		for _, candidate := range candidates {
			result.Scanned++

			loc := locations.get(candidate.Timezone)

			for _, slot := range medicationSlots(candidate, cmd.From, cmd.To, loc) {
				created, err := h.saveReminder(ctx, buildMedicationNotification(candidate, slot, loc))
				if err != nil {
					return err
				}

				if created {
					result.MedicationCreated++
				} else {
					result.Skipped++
				}
			}
		}

		last := candidates[len(candidates)-1]
		afterMedicationID = last.MedicationID
		afterMedicationTimeID = last.MedicationTimeID

		if len(candidates) < limit {
			return nil
		}
	}
}

func (h *GenerateDueRemindersHandler) generateAppointmentReminders(
	ctx context.Context,
	cmd command.GenerateDueRemindersCommand,
	locations *locationCache,
	result *dto.ReminderRunResultDTO,
) error {
	limit := cmd.Limit
	if limit <= 0 {
		limit = 500
	}

	// A reminder at time T is for a visit one lead-time later.
	from := cmd.From.Add(appointmentLeadTime)
	to := cmd.To.Add(appointmentLeadTime)

	// Paginated for the same reason the medication scan is: a single LIMIT
	// silently dropped every visit past the batch size, and because the window
	// moves with the clock those visits were never picked up on a later tick
	// either — the patient simply got no reminder.
	var afterVisitDate time.Time
	var afterVisitID int64

	for {
		candidates, err := h.reminders.FindAppointmentCandidates(
			ctx, from, to, limit, afterVisitDate, afterVisitID,
		)
		if err != nil {
			return fmt.Errorf("find appointment candidates: %w", err)
		}

		if len(candidates) == 0 {
			return nil
		}

		if err := h.processAppointmentCandidates(ctx, candidates, locations, result); err != nil {
			return err
		}

		last := candidates[len(candidates)-1]
		afterVisitDate = last.VisitDate
		afterVisitID = last.VisitID

		if len(candidates) < limit {
			return nil
		}
	}
}

func (h *GenerateDueRemindersHandler) processAppointmentCandidates(
	ctx context.Context,
	candidates []readmodel.DueAppointmentCandidate,
	locations *locationCache,
	result *dto.ReminderRunResultDTO,
) error {
	for _, candidate := range candidates {
		result.Scanned++

		loc := locations.get(candidate.Timezone)
		localDate := candidate.VisitDate.In(loc).Format("2006-01-02")

		created, err := h.saveReminder(ctx, buildAppointmentNotification(candidate, localDate))
		if err != nil {
			return err
		}

		if created {
			result.AppointmentCreated++
		} else {
			result.Skipped++
		}
	}

	return nil
}

func (h *GenerateDueRemindersHandler) saveReminder(
	ctx context.Context,
	n *entities.Notification,
) (bool, error) {
	created, err := h.notifications.SaveIfNotExists(ctx, n)
	if err != nil {
		return false, fmt.Errorf("save reminder %q: %w", derefString(n.DedupeKey), err)
	}

	return created, nil
}

// medicationSlots expands one (medication, time-of-day) pair into the concrete
// UTC instants that fall inside [from, to].
func medicationSlots(
	c readmodel.DueMedicationCandidate,
	from, to time.Time,
	loc *time.Location,
) []time.Time {
	hour, minute, ok := parseTimeOfDay(c.TimeOfDay)
	if !ok {
		return nil
	}

	var slots []time.Time

	// Walk local calendar days at local midnight, so every comparison below is
	// day-to-day. Starting one day early covers a local day that begins before
	// `from` in UTC; ending one late covers the mirror case.
	day := startOfLocalDay(from.In(loc)).AddDate(0, 0, -1)
	last := startOfLocalDay(to.In(loc)).AddDate(0, 0, 1)

	for !day.After(last) {
		if dayMatchesFrequency(c, day, loc) {
			// time.Date resolves DST gaps and overlaps deterministically: a
			// nonexistent local time normalises forward, an ambiguous one takes
			// the first offset. Combined with the dedupe key that yields at most
			// one reminder per local day per slot.
			slot := time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, loc).UTC()

			if !slot.Before(from) && !slot.After(to) {
				slots = append(slots, slot)
			}
		}

		day = day.AddDate(0, 0, 1)
	}

	return slots
}

// dayMatchesFrequency decides whether a medication is taken on a given local
// day, which arrives as local midnight. medication_times already says *which
// times*, so frequency only gates which days.
func dayMatchesFrequency(c readmodel.DueMedicationCandidate, day time.Time, loc *time.Location) bool {
	if c.StartDate != nil && day.Before(startOfLocalDay(c.StartDate.In(loc))) {
		return false
	}

	// Compared at day granularity so the course still runs on its end date
	// rather than stopping the day before.
	if c.EndDate != nil && day.After(startOfLocalDay(c.EndDate.In(loc))) {
		return false
	}

	frequency := valueobjects.FrequencyDaily
	if c.Frequency != nil && strings.TrimSpace(*c.Frequency) != "" {
		frequency = valueobjects.MedicationFrequency(strings.TrimSpace(*c.Frequency))
	}

	switch frequency {
	case valueobjects.FrequencyAsNeeded:
		// Taken on demand — a scheduled reminder would be noise.
		return false

	case valueobjects.FrequencyOnce:
		if c.StartDate == nil {
			return false
		}
		return sameLocalDay(day, c.StartDate.In(loc))

	case valueobjects.FrequencyWeekly:
		if c.StartDate == nil {
			return false
		}
		return day.Weekday() == c.StartDate.In(loc).Weekday()

	default:
		// daily, twice_daily, three_times_daily and anything unrecognised:
		// every day, with medication_times supplying the per-day times.
		return true
	}
}

func buildMedicationNotification(
	c readmodel.DueMedicationCandidate,
	slot time.Time,
	loc *time.Location,
) *entities.Notification {
	// slot is UTC; the dedupe key is scoped to the user's *local* day, so the
	// location has to be supplied rather than read back off the instant.
	localDate := slot.In(loc).Format("2006-01-02")

	// The dedupe key is scoped to the user's local day, which is what a person
	// means by "my 8am dose today".
	dedupeKey := fmt.Sprintf("med:%d:%d:%s", c.MedicationID, c.MedicationTimeID, localDate)

	body := c.TimeOfDay
	if c.Dosage != nil && strings.TrimSpace(*c.Dosage) != "" {
		body = fmt.Sprintf("%s · %s", strings.TrimSpace(*c.Dosage), c.TimeOfDay)
	}

	channel := notificationChannelInApp

	return &entities.Notification{
		UserID:      c.UserID,
		Title:       fmt.Sprintf("Time for %s", c.MedicationName),
		Body:        body,
		Type:        notificationTypeMedication,
		Channel:     &channel,
		Status:      notificationStatusUnread,
		ScheduledAt: &slot,
		// There is no push transport: "sent" means "visible in the inbox".
		SentAt:    &slot,
		DedupeKey: &dedupeKey,
		Metadata: map[string]any{
			"kind":               notificationTypeMedication,
			"medication_id":      c.MedicationID,
			"medication_time_id": c.MedicationTimeID,
			"local_date":         localDate,
		},
		CreatedAt: slot,
	}
}

func buildAppointmentNotification(
	c readmodel.DueAppointmentCandidate,
	localDate string,
) *entities.Notification {
	dedupeKey := fmt.Sprintf("visit:%d:%s:24h", c.VisitID, localDate)

	title := "Upcoming appointment"
	if c.HospitalName != nil && strings.TrimSpace(*c.HospitalName) != "" {
		title = fmt.Sprintf("Appointment at %s", strings.TrimSpace(*c.HospitalName))
	}

	body := "Tomorrow"
	if c.Doctor != nil && strings.TrimSpace(*c.Doctor) != "" {
		body = fmt.Sprintf("Tomorrow with %s", strings.TrimSpace(*c.Doctor))
	}

	scheduledAt := c.VisitDate.Add(-appointmentLeadTime)
	channel := notificationChannelInApp

	return &entities.Notification{
		UserID:      c.UserID,
		Title:       title,
		Body:        body,
		Type:        notificationTypeAppointment,
		Channel:     &channel,
		Status:      notificationStatusUnread,
		ScheduledAt: &scheduledAt,
		SentAt:      &scheduledAt,
		DedupeKey:   &dedupeKey,
		Metadata: map[string]any{
			"kind":       notificationTypeAppointment,
			"visit_id":   c.VisitID,
			"local_date": localDate,
			"visit_date": c.VisitDate.UTC().Format(time.RFC3339),
		},
		CreatedAt: scheduledAt,
	}
}

// locationCache resolves IANA names once per tick and never fails a run over a
// single bad row — one unparseable timezone must not stop everyone's reminders.
type locationCache struct {
	cache  map[string]*time.Location
	logger *zap.Logger
}

func newLocationCache(logger *zap.Logger) *locationCache {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &locationCache{cache: make(map[string]*time.Location), logger: logger}
}

func (c *locationCache) get(name string) *time.Location {
	name = strings.TrimSpace(name)
	if name == "" {
		return time.UTC
	}

	if loc, ok := c.cache[name]; ok {
		return loc
	}

	loc, err := time.LoadLocation(name)
	if err != nil {
		c.logger.Warn("unresolvable user timezone, falling back to UTC",
			zap.String("timezone", name), zap.Error(err))
		loc = time.UTC
	}

	c.cache[name] = loc

	return loc
}

func parseTimeOfDay(value string) (hour, minute int, ok bool) {
	parsed, err := time.Parse("15:04", strings.TrimSpace(value))
	if err != nil {
		return 0, 0, false
	}
	return parsed.Hour(), parsed.Minute(), true
}

func startOfLocalDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func sameLocalDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
