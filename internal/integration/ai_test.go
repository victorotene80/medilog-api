package integration

import (
	"testing"
)

func TestIntegration_AIConversation_Create(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	body := map[string]interface{}{
		"title": "Medication Questions",
	}

	resp, result := ts.doRequest("POST", "/api/v1/ai/conversations", body, ts.authHeaders(accessToken))

	if resp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_AIConversation_List(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	createBody := map[string]interface{}{
		"title": "Health Questions",
	}
	ts.doRequest("POST", "/api/v1/ai/conversations", createBody, ts.authHeaders(accessToken))

	resp, result := ts.doRequest("GET", "/api/v1/ai/conversations", nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_AIConversation_Get(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	createBody := map[string]interface{}{
		"title": "Medication Inquiry",
	}
	_, createResult := ts.doRequest("POST", "/api/v1/ai/conversations", createBody, ts.authHeaders(accessToken))

	publicID := ts.getPublicID(createResult)

	resp, result := ts.doRequest("GET", "/api/v1/ai/conversations/"+publicID, nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_AIConversation_Archive(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	createBody := map[string]interface{}{
		"title": "To Archive",
	}
	_, createResult := ts.doRequest("POST", "/api/v1/ai/conversations", createBody, ts.authHeaders(accessToken))

	publicID := ts.getPublicID(createResult)

	resp, result := ts.doRequest("PATCH", "/api/v1/ai/conversations/"+publicID+"/archive", nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_AIConversation_Unauthenticated(t *testing.T) {
	ts := setupTestServer(t)

	resp, _ := ts.doRequest("GET", "/api/v1/ai/conversations", nil, nil)

	if resp.StatusCode != 401 {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}
