package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/handlers"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence"
	"gorm.io/gorm"
)

// The scheduler ticker is disabled in tests (SCHEDULER_ENABLED=false); these
// drive the handler directly against the real database so the window is fixed
// and the assertions cannot race a background tick.
func newReminderHandler(t *testing.T, db *gorm.DB) *handlers.GenerateDueRemindersHandler {
	t.Helper()

	return handlers.NewGenerateDueRemindersHandler(
		persistence.NewReminderRepository(db),
		persistence.NewNotificationRepository(db),
		nil,
		func() time.Time { return time.Now().UTC() },
	)
}

func openIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set — skipping integration tests")
	}

	host, port, user, pass, dbname := parseDBURL(dbURL)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbname)

	db, err := openTestDB(dsn)
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}

	t.Cleanup(func() { closeTestDB(db) })

	return db
}

// seedMedicationWithTime inserts a medication and one daily time-of-day for a
// user, mirroring how CreateMedicationHandler stores medication_times: a
// wall-clock time parsed with layout "15:04", which yields a zero date at UTC.
func seedMedicationWithTime(
	t *testing.T,
	db *gorm.DB,
	userID int64,
	name, timeOfDay string,
) int64 {
	t.Helper()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}

	var medicationID int64
	err = sqlDB.QueryRow(`
		INSERT INTO medications (user_id, name, frequency, is_completed, start_date, created_at, updated_at)
		VALUES ($1, $2, 'daily', false, NOW() - INTERVAL '1 day', NOW(), NOW())
		RETURNING id`, userID, name).Scan(&medicationID)
	if err != nil {
		t.Fatalf("insert medication: %v", err)
	}

	parsed, err := time.Parse("15:04", timeOfDay)
	if err != nil {
		t.Fatalf("bad time of day %q: %v", timeOfDay, err)
	}

	if _, err := sqlDB.Exec(`
		INSERT INTO medication_times (medication_id, time_value, created_at)
		VALUES ($1, $2, NOW())`, medicationID, parsed); err != nil {
		t.Fatalf("insert medication time: %v", err)
	}

	return medicationID
}

func countReminders(t *testing.T, db *gorm.DB, userID int64) int {
	t.Helper()

	sqlDB, _ := db.DB()

	var count int
	if err := sqlDB.QueryRow(`
		SELECT COUNT(*) FROM notifications
		 WHERE user_id = $1 AND type = 'medication_reminder' AND deleted_at IS NULL`,
		userID).Scan(&count); err != nil {
		t.Fatalf("count reminders: %v", err)
	}

	return count
}

// The whole reason the dedupe key exists: the scheduler re-scans an overlapping
// window every tick, so generating twice must not produce two notifications.
// Without the unique index this test yields 2.
func TestIntegration_Scheduler_IsIdempotent(t *testing.T) {
	ts := setupTestServer(t)
	db := openIntegrationDB(t)

	email := randomTestEmail()
	_, registerResult := ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	userID := parseInt64(ts.getUserID(registerResult))

	// A slot inside the window below.
	slot := time.Now().UTC().Add(-30 * time.Minute)
	seedMedicationWithTime(t, db, userID, "Amoxicillin", slot.Format("15:04"))

	handler := newReminderHandler(t, db)

	cmd := command.GenerateDueRemindersCommand{
		From:  time.Now().UTC().Add(-2 * time.Hour),
		To:    time.Now().UTC().Add(5 * time.Minute),
		Limit: 500,
	}

	first, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	assertReminderRun(t, first)

	// MedicationCreated counts the whole scan, not just this user: every test in
	// the package shares one database, and any other medication whose time falls
	// in the two-hour window above is picked up by the same run. Requiring
	// exactly 1 here made the test depend on what the rest of the suite had
	// already created — and on the wall-clock time it ran at. The per-user row
	// count immediately below is the assertion that actually pins idempotency.
	if first.MedicationCreated < 1 {
		t.Fatalf("expected this user's reminder to be created, got %d (%+v)",
			first.MedicationCreated, first)
	}
	if got := countReminders(t, db, userID); got != 1 {
		t.Fatalf("expected 1 notification row after the first run, got %d", got)
	}

	second, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	assertReminderRun(t, second)

	if second.MedicationCreated != 0 {
		t.Fatalf("the second run must create nothing, got %d (%+v)",
			second.MedicationCreated, second)
	}
	if second.Skipped < 1 {
		t.Fatalf("the second run should have skipped the existing slot, got %+v", second)
	}
	if got := countReminders(t, db, userID); got != 1 {
		t.Fatalf("re-running must not duplicate: expected 1 notification, got %d", got)
	}
}

// A dismissed reminder must stay dismissed rather than reappearing next tick.
// This is why the dedupe index deliberately does not exclude soft-deleted rows.
func TestIntegration_Scheduler_DoesNotResurrectDeletedReminders(t *testing.T) {
	ts := setupTestServer(t)
	db := openIntegrationDB(t)

	email := randomTestEmail()
	_, registerResult := ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	userID := parseInt64(ts.getUserID(registerResult))

	slot := time.Now().UTC().Add(-30 * time.Minute)
	seedMedicationWithTime(t, db, userID, "Amoxicillin", slot.Format("15:04"))

	handler := newReminderHandler(t, db)
	cmd := command.GenerateDueRemindersCommand{
		From:  time.Now().UTC().Add(-2 * time.Hour),
		To:    time.Now().UTC().Add(5 * time.Minute),
		Limit: 500,
	}

	if _, err := handler.Handle(context.Background(), cmd); err != nil {
		t.Fatalf("first run failed: %v", err)
	}

	sqlDB, _ := db.DB()
	if _, err := sqlDB.Exec(
		`UPDATE notifications SET deleted_at = NOW() WHERE user_id = $1`, userID); err != nil {
		t.Fatalf("soft delete: %v", err)
	}

	result, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	if result.MedicationCreated != 0 {
		t.Fatalf("a dismissed reminder must not be regenerated, got %+v", result)
	}
	if got := countReminders(t, db, userID); got != 0 {
		t.Fatalf("expected the reminder to stay dismissed, found %d live rows", got)
	}
}

// Users who turn medication reminders off must generate nothing — the toggle
// has to actually do something.
func TestIntegration_Scheduler_HonoursDisabledPreference(t *testing.T) {
	ts := setupTestServer(t)
	db := openIntegrationDB(t)

	email := randomTestEmail()
	_, registerResult := ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	userID := parseInt64(ts.getUserID(registerResult))
	_, loginResult := ts.loginUser(email, "Password123!")
	token := ts.getAccessToken(loginResult)

	slot := time.Now().UTC().Add(-30 * time.Minute)
	seedMedicationWithTime(t, db, userID, "Amoxicillin", slot.Format("15:04"))

	resp, body := ts.doRequest("PATCH", "/api/v1/users/me/notification-preferences",
		map[string]interface{}{"medication_reminders_enabled": false},
		ts.authHeaders(token))
	if resp.StatusCode != 200 {
		t.Fatalf("could not disable reminders: %d %v", resp.StatusCode, body)
	}

	result, err := newReminderHandler(t, db).Handle(context.Background(),
		command.GenerateDueRemindersCommand{
			From:  time.Now().UTC().Add(-2 * time.Hour),
			To:    time.Now().UTC().Add(5 * time.Minute),
			Limit: 500,
		})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if result.MedicationCreated != 0 {
		t.Fatalf("reminders were disabled but %d were created", result.MedicationCreated)
	}
	if got := countReminders(t, db, userID); got != 0 {
		t.Fatalf("expected no notifications, got %d", got)
	}
}

// Generated reminders must be readable through the inbox the client polls.
func TestIntegration_Scheduler_ReminderAppearsInInbox(t *testing.T) {
	ts := setupTestServer(t)
	db := openIntegrationDB(t)

	email := randomTestEmail()
	_, registerResult := ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	userID := parseInt64(ts.getUserID(registerResult))
	_, loginResult := ts.loginUser(email, "Password123!")
	token := ts.getAccessToken(loginResult)

	slot := time.Now().UTC().Add(-30 * time.Minute)
	seedMedicationWithTime(t, db, userID, "Amoxicillin", slot.Format("15:04"))

	if _, err := newReminderHandler(t, db).Handle(context.Background(),
		command.GenerateDueRemindersCommand{
			From:  time.Now().UTC().Add(-2 * time.Hour),
			To:    time.Now().UTC().Add(5 * time.Minute),
			Limit: 500,
		}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	resp, result := ts.doRequest("GET", "/api/v1/notifications", nil, ts.authHeaders(token))
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 from the inbox, got %d: %v", resp.StatusCode, result)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected inbox shape: %v", result)
	}

	items, ok := data["notifications"].([]interface{})
	if !ok || len(items) == 0 {
		t.Fatalf("expected at least one notification in the inbox, got: %v", data)
	}

	first, _ := items[0].(map[string]interface{})
	if first["type"] != "medication_reminder" {
		t.Fatalf("expected a medication_reminder, got %v", first["type"])
	}
	if first["title"] != "Time for Amoxicillin" {
		t.Fatalf("unexpected title: %v", first["title"])
	}

	// The id must be usable against the detail route — it used to be a UUID
	// handed to an endpoint that parsed an int64.
	id, _ := first["id"].(string)
	if id == "" {
		t.Fatalf("notification has no id: %v", first)
	}

	detailResp, detailResult := ts.doRequest(
		"GET", "/api/v1/notifications/"+id, nil, ts.authHeaders(token))
	if detailResp.StatusCode != 200 {
		t.Fatalf("expected 200 fetching the notification by its own id, got %d: %v",
			detailResp.StatusCode, detailResult)
	}
}

func assertReminderRun(t *testing.T, result *dto.ReminderRunResultDTO) {
	t.Helper()

	if result == nil {
		t.Fatal("nil reminder run result")
	}
	if result.LockNotAcquired {
		t.Fatal("advisory lock was unexpectedly held by another process")
	}
}
