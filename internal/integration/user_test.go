package integration

import (
	"testing"
)

func TestIntegration_User_GetMe(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	resp, result := ts.doRequest("GET", "/api/v1/users/me", nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_User_GetMe_Unauthenticated(t *testing.T) {
	ts := setupTestServer(t)

	resp, _ := ts.doRequest("GET", "/api/v1/users/me", nil, nil)

	if resp.StatusCode != 401 {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestIntegration_EmergencyContact_Create(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	body := map[string]interface{}{
		"name":         "Jane Doe",
		"relationship": "Mother",
		"phone":        "+2348012345678",
		"country_code": "NG",
		"is_primary":   true,
	}

	resp, result := ts.doRequest("POST", "/api/v1/emergency-contacts", body, ts.authHeaders(accessToken))

	if resp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_EmergencyContact_CreateUnauthenticated(t *testing.T) {
	ts := setupTestServer(t)

	body := map[string]interface{}{
		"name":         "Jane Doe",
		"relationship": "Mother",
		"phone":        "+2348012345678",
		"country_code": "NG",
		"is_primary":   true,
	}

	resp, _ := ts.doRequest("POST", "/api/v1/emergency-contacts", body, nil)

	if resp.StatusCode != 401 {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}
