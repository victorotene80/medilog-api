package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/victorotene80/medilog-api/internal/bootstrap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type testServer struct {
	*httptest.Server
	t  *testing.T
	db string
}

func parseDBURL(rawURL string) (host, port, user, password, dbname string) {
	u, _ := url.Parse(rawURL)
	host = u.Hostname()
	port = u.Port()
	if port == "" {
		port = "5432"
	}
	user = u.User.Username()
	password, _ = u.User.Password()
	dbname = strings.TrimPrefix(u.Path, "/")
	return
}

func flushRateLimitKeys(t *testing.T) {
	t.Helper()

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPass := os.Getenv("REDIS_PASSWORD")

	if redisHost == "" {
		redisHost = "localhost"
	}
	if redisPort == "" {
		redisPort = "6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPass,
		DB:       0,
	})
	defer rdb.Close()

	ctx := t.Context()
	keys, err := rdb.Keys(ctx, "rate_limit:*").Result()
	if err == nil && len(keys) > 0 {
		rdb.Del(ctx, keys...)
	}
}

func setupTestServer(t *testing.T) *testServer {
	t.Helper()

	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set — skipping integration tests")
	}

	host, port, user, pass, dbname := parseDBURL(dbURL)

	RequiredEnvVars := map[string]string{
		"SESSION_PEPPER":             "test-pepper-3d7c1e94e5f8452c90bbd03fa9d44b0dfc68a157be2cf49162d7ee9148b7e32d",
		"JWT_SECRET":                 "test-jwt-secret-cf97f3b0b8a54e0c9131ddf482ca38d22bc69c7e90ead53c4f4b374cd682e75f0ad73d6663d417e8bb23ac956f98ef1c79b4f9c5b4ccfbb7bc3f9f3e0b16db72",
		"ACCESS_TOKEN_TTL":           "15m",
		"REFRESH_TOKEN_TTL":          "168h",
		"DB_HOST":                    host,
		"DB_PORT":                    port,
		"DB_USER":                    user,
		"DB_PASSWORD":                pass,
		"DB_NAME":                    dbname,
		"DB_SSLMODE":                 "disable",
		"DB_MAX_OPEN":                "20",
		"DB_MAX_IDLE":                "10",
		"DB_MAX_LIFETIME":            "300",
		"REDIS_HOST":                 "localhost",
		"REDIS_PORT":                 "6379",
		"REDIS_PASSWORD":             "redis",
		"REDIS_DB":                   "0",
		"REDIS_TTL":                  "15m",
		"GOOGLE_CLIENT_ID":           "test-google-client-id",
		"AI_CONTEXT_WINDOW_TOKENS":   "3500",
		"AI_MAX_CONTEXT_MESSAGES":    "20",
		"AI_SUMMARY_TOKEN_THRESHOLD": "6000",
		"SMS_MAX_BULK_FAILURES":      "3",
		"BULK_SMS_API_TOKEN":         "test-token",
		"BULK_SMS_SENDER":            "Medilog",
		"TWILIO_ACCOUNT_SID":         "test-sid",
		"TWILIO_AUTH_TOKEN":          "test-token",
		"TELNYX_API_KEY":             "test-key",
		"SMS_ENABLED":                "false",
		"AI_ENABLED":                 "false",
		"MESSAGING_ENABLED":          "false",
		"OTEL_ENABLED":               "false",
		"DB_AUTO_MIGRATE":            "false",
		// The reminder ticker must not race the assertions; scheduler tests
		// invoke the handler directly instead.
		"SCHEDULER_ENABLED": "false",
		"DB_VERIFY_SCHEMA":  "true",
	}

	for k, v := range RequiredEnvVars {
		if os.Getenv(k) == "" {
			t.Setenv(k, v)
		}
	}

	flushRateLimitKeys(t)

	app, err := bootstrap.InitializeApp()
	if err != nil {
		t.Fatalf("failed to initialize app: %v", err)
	}

	// Stop drains the background workers; Close returns the database and Redis
	// pools. Without Close every test in the package leaks DB_MAX_OPEN
	// connections for the whole run, and the suite eventually dies part-way
	// through with "sorry, too many clients already" — a failure that lands on
	// whichever test happened to be running when the server hit its limit.
	t.Cleanup(func() {
		app.Stop()
		app.Close()
	})

	ts := httptest.NewServer(app.Router)
	t.Cleanup(ts.Close)

	return &testServer{Server: ts, t: t, db: dbname}
}

func (s *testServer) registerAndCompleteOnboarding(email, password, firstName, lastName string) (*http.Response, map[string]interface{}) {
	s.t.Helper()

	resp, result := s.registerUser(email, password, firstName, lastName)

	if resp.StatusCode == 201 {
		userID := s.getUserID(result)
		completeOnboardingForUser(s.t, s.db, userID)
	}

	return resp, result
}

func (s *testServer) doRequest(method, path string, body interface{}, headers map[string]string) (*http.Response, map[string]interface{}) {
	s.t.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			s.t.Fatalf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, s.URL+path, reqBody)
	if err != nil {
		s.t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.t.Fatalf("request failed: %v", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		s.t.Fatalf("failed to read response body: %v", err)
	}
	resp.Body.Close()

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		result = map[string]interface{}{
			"raw": string(respBody),
		}
	}

	return resp, result
}

func (s *testServer) registerUser(email, password, firstName, lastName string) (*http.Response, map[string]interface{}) {
	s.t.Helper()

	body := map[string]interface{}{
		"email":        email,
		"password":     password,
		"first_name":   firstName,
		"last_name":    lastName,
		"dob":          "1990-01-01",
		"sex":          1,
		"blood_type":   "O+",
		"country_code": "NG",
	}

	return s.doRequest("POST", "/api/v1/auth/register", body, nil)
}

func (s *testServer) loginUser(email, password string) (*http.Response, map[string]interface{}) {
	s.t.Helper()

	body := map[string]interface{}{
		"email":    email,
		"password": password,
	}

	return s.doRequest("POST", "/api/v1/auth/login", body, nil)
}

func (s *testServer) getAccessToken(result map[string]interface{}) string {
	s.t.Helper()

	if data, ok := result["data"].(map[string]interface{}); ok {
		if tokens, ok := data["tokens"].(map[string]interface{}); ok {
			if at, ok := tokens["access_token"].(string); ok {
				return at
			}
		}
		if at, ok := data["access_token"].(string); ok {
			return at
		}
		if atObj, ok := data["access_token"].(map[string]interface{}); ok {
			if v, ok := atObj["Value"].(string); ok {
				return v
			}
		}
	}

	if at, ok := result["access_token"].(string); ok {
		return at
	}
	if atObj, ok := result["access_token"].(map[string]interface{}); ok {
		if v, ok := atObj["Value"].(string); ok {
			return v
		}
	}

	s.t.Fatalf("could not extract access_token from: %v", result)
	return ""
}

func (s *testServer) getRefreshToken(result map[string]interface{}) string {
	s.t.Helper()

	if data, ok := result["data"].(map[string]interface{}); ok {
		if tokens, ok := data["tokens"].(map[string]interface{}); ok {
			if rt, ok := tokens["refresh_token"].(string); ok {
				return rt
			}
		}
		if rt, ok := data["refresh_token"].(string); ok {
			return rt
		}
	}

	if rt, ok := result["refresh_token"].(string); ok {
		return rt
	}

	s.t.Fatalf("could not extract refresh_token from: %v", result)
	return ""
}

func (s *testServer) getUserID(result map[string]interface{}) string {
	s.t.Helper()

	if data, ok := result["data"].(map[string]interface{}); ok {
		if id, ok := data["user_id"].(string); ok {
			return id
		}
		if id, ok := data["user_id"].(float64); ok {
			return fmt.Sprintf("%.0f", id)
		}
	}

	s.t.Fatalf("could not extract user id from: %v", result)
	return ""
}

func (s *testServer) getPublicID(result map[string]interface{}) string {
	s.t.Helper()

	if data, ok := result["data"].(map[string]interface{}); ok {
		if id, ok := data["public_id"].(string); ok {
			return id
		}
		if id, ok := data["id"].(string); ok {
			return id
		}
	}

	s.t.Fatalf("could not extract public_id from: %v", result)
	return ""
}

func (s *testServer) authHeaders(accessToken string) map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}
}

// testEmailSeq disambiguates addresses generated within the same clock tick.
// time.Now() is not guaranteed to advance between two adjacent calls, so a
// test that asked for two addresses in a row could get the same one twice —
// the second registration then failed as a duplicate and the test silently
// operated on a single account, failing later and only sometimes.
var testEmailSeq atomic.Uint64

func randomTestEmail() string {
	return fmt.Sprintf("test-%d-%d-%d@example.com",
		os.Getpid(), time.Now().UnixNano(), testEmailSeq.Add(1))
}

func openTestDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func closeTestDB(db *gorm.DB) {
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}
}

func completeOnboardingForUser(t *testing.T, dbName string, userID string) {
	t.Helper()

	dbURL := os.Getenv("TEST_DB_URL")
	host, port, user, pass, _ := parseDBURL(dbURL)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbName)

	db, err := openTestDB(dsn)
	if err != nil {
		t.Fatalf("failed to open DB for onboarding update: %v", err)
	}
	defer closeTestDB(db)

	sqlDB, _ := db.DB()
	_, err = sqlDB.Exec("UPDATE users SET is_onboarding_completed = true, status = 'active' WHERE id = $1", userID)
	if err != nil {
		t.Fatalf("failed to mark onboarding completed: %v", err)
	}
}

// activateUserWithoutOnboarding marks a user verified but leaves onboarding
// incomplete — the state a real user is in right after OTP verification and
// before they add their first emergency contact.
//
// This is exactly the state that used to be locked out: the only endpoint that
// completes onboarding sat behind a middleware requiring onboarding to already
// be complete.
func activateUserWithoutOnboarding(t *testing.T, dbName string, userID string) {
	t.Helper()

	dbURL := os.Getenv("TEST_DB_URL")
	host, port, user, pass, _ := parseDBURL(dbURL)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbName)

	db, err := openTestDB(dsn)
	if err != nil {
		t.Fatalf("failed to open DB for activation: %v", err)
	}
	defer closeTestDB(db)

	sqlDB, _ := db.DB()
	_, err = sqlDB.Exec(
		"UPDATE users SET status = 'active', is_onboarding_completed = false WHERE id = $1",
		userID,
	)
	if err != nil {
		t.Fatalf("failed to activate user: %v", err)
	}
}

func randomTestPhone() string {
	return fmt.Sprintf("+1555%07d", time.Now().UnixNano()%10000000)
}

func parseInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func cleanupAllTables(db *gorm.DB) {
	// These names must match the real schema. Several previously listed here
	// (otps, feedbacks, refresh_tokens, onboarding_progress, drug_scan_results)
	// do not exist, and TRUNCATE on a missing table fails silently through
	// db.Exec — so those tables were never actually being cleaned between runs.
	tables := []string{
		"users", "user_profiles", "user_auth_providers", "emergency_contacts",
		"refresh_token", "otp_codes",
		"visits", "user_allergies", "drug_scans", "fun_facts",
		"medications", "medication_times", "medication_adherence_logs",
		"registered_medicines", "feedback", "support_tickets",
		"support_messages", "support_attachments", "ai_conversations",
		"ai_messages", "notifications", "outbox_events", "audit_logs",
	}
	for _, table := range tables {
		if err := db.Exec("TRUNCATE TABLE " + table + " CASCADE").Error; err != nil {
			// Loud rather than silent: a rename here otherwise leaves state
			// bleeding between tests with no signal at all.
			fmt.Printf("warning: could not truncate %q: %v\n", table, err)
		}
	}
}

func TestMain(m *testing.M) {
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		os.Exit(m.Run())
	}

	host, port, user, pass, dbname := parseDBURL(dbURL)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)

	db, err := openTestDB(dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open test DB: %v\n", err)
		os.Exit(1)
	}
	defer closeTestDB(db)

	cleanupAllTables(db)

	os.Exit(m.Run())
}
