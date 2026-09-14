package integration

import (
	"testing"
)

func TestIntegration_Visit_Create(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	body := map[string]interface{}{
		"hospital_name": "Lagos University Teaching Hospital",
		"diagnosis":     "Malaria",
		"visit_date":    "2026-08-29T10:00:00Z",
		"outcome":       "Resolved",
		"doctor":        "Dr. Smith",
	}

	resp, result := ts.doRequest("POST", "/api/v1/visits", body, ts.authHeaders(accessToken))

	if resp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Visit_CreateValidation(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	body := map[string]interface{}{
		"hospital_name": "Test Hospital",
	}

	resp, _ := ts.doRequest("POST", "/api/v1/visits", body, ts.authHeaders(accessToken))

	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestIntegration_Visit_List(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	createBody := map[string]interface{}{
		"hospital_name": "General Hospital",
		"diagnosis":     "Checkup",
		"visit_date":    "2026-08-29T10:00:00Z",
	}
	ts.doRequest("POST", "/api/v1/visits", createBody, ts.authHeaders(accessToken))

	resp, result := ts.doRequest("GET", "/api/v1/visits", nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Visit_ListDateRange(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	createBody := map[string]interface{}{
		"hospital_name": "General Hospital",
		"diagnosis":     "Checkup",
		"visit_date":    "2026-08-29T10:00:00Z",
	}
	ts.doRequest("POST", "/api/v1/visits", createBody, ts.authHeaders(accessToken))

	resp, result := ts.doRequest("GET", "/api/v1/visits?from=2026-01-01&to=2026-12-31", nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Visit_ListInvalidDateRange(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	resp, result := ts.doRequest("GET", "/api/v1/visits?from=2026-01-01", nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Visit_Get(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	createBody := map[string]interface{}{
		"hospital_name": "City Hospital",
		"diagnosis":     "Flu",
		"visit_date":    "2026-08-29T10:00:00Z",
	}
	ts.doRequest("POST", "/api/v1/visits", createBody, ts.authHeaders(accessToken))

	listResp, listResult := ts.doRequest("GET", "/api/v1/visits", nil, ts.authHeaders(accessToken))
	if listResp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", listResp.StatusCode)
	}

	listData, ok := listResult["data"].([]interface{})
	if !ok || len(listData) == 0 {
		t.Fatalf("expected visits in list, got: %v", listResult)
	}

	firstVisit, ok := listData[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected visit object, got: %v", listData[0])
	}

	publicID, ok := firstVisit["public_id"].(string)
	if !ok {
		t.Fatalf("expected public_id string, got: %v", firstVisit)
	}

	resp, result := ts.doRequest("GET", "/api/v1/visits/"+publicID, nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Visit_Unauthenticated(t *testing.T) {
	ts := setupTestServer(t)

	resp, _ := ts.doRequest("GET", "/api/v1/visits", nil, nil)

	if resp.StatusCode != 401 {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}
