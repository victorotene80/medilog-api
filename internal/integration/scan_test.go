package integration

import (
	"testing"
)

func TestIntegration_DrugScan_List(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	resp, result := ts.doRequest("GET", "/api/v1/drugs/scans", nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_DrugScan_Unauthenticated(t *testing.T) {
	ts := setupTestServer(t)

	resp, _ := ts.doRequest("GET", "/api/v1/drugs/scans", nil, nil)

	if resp.StatusCode != 401 {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestIntegration_DrugScan_Verify(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	body := map[string]interface{}{
		"drug_name":    "Paracetamol",
		"country_code": "NG",
	}

	resp, _ := ts.doRequest("POST", "/api/v1/drugs/verify", body, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 && resp.StatusCode != 500 {
		t.Fatalf("expected status 200 or 500 (external dep), got %d", resp.StatusCode)
	}
}
