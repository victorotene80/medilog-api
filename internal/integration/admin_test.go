package integration

import (
	"fmt"
	"os"
	"testing"
)

func TestIntegration_AuditLog_NonAdminDenied(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	resp, _ := ts.doRequest("GET", "/api/v1/audit-logs", nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 403 {
		t.Fatalf("expected status 403, got %d", resp.StatusCode)
	}
}

func TestIntegration_AuditLog_Unauthenticated(t *testing.T) {
	ts := setupTestServer(t)

	resp, _ := ts.doRequest("GET", "/api/v1/audit-logs", nil, nil)

	if resp.StatusCode != 401 {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestIntegration_AuditLog_AdminAccess(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	_, regResult := ts.registerAndCompleteOnboarding(email, "Password123!", "Admin", "User")
	userID := ts.getUserID(regResult)

	setUserRoleAsAdmin(t, ts.db, userID)

	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	resp, result := ts.doRequest("GET", "/api/v1/audit-logs?limit=10&offset=0", nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func setUserRoleAsAdmin(t *testing.T, dbName string, userID string) {
	t.Helper()

	dbURL := os.Getenv("TEST_DB_URL")
	host, port, user, pass, _ := parseDBURL(dbURL)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbName)

	db, err := openTestDB(dsn)
	if err != nil {
		t.Fatalf("failed to open DB for admin update: %v", err)
	}
	defer closeTestDB(db)

	sqlDB, _ := db.DB()
	_, err = sqlDB.Exec("UPDATE users SET role = 'admin' WHERE id = $1", userID)
	if err != nil {
		t.Fatalf("failed to set admin role: %v", err)
	}
}
