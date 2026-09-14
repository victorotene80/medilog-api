package integration

import (
	"testing"
)

func TestIntegration_SupportTicket_Create(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	body := map[string]interface{}{
		"subject":  "Test Support Ticket",
		"category": "general",
		"priority": "normal",
		"message":  "This is a test support ticket message.",
	}

	resp, result := ts.doRequest("POST", "/api/v1/support/tickets", body, ts.authHeaders(accessToken))

	if resp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_SupportTicket_CreateUnauthenticated(t *testing.T) {
	ts := setupTestServer(t)

	body := map[string]interface{}{
		"subject":  "Test Support Ticket",
		"category": "general",
		"priority": "normal",
		"message":  "This is a test support ticket message.",
	}

	resp, _ := ts.doRequest("POST", "/api/v1/support/tickets", body, nil)

	if resp.StatusCode != 401 {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestIntegration_SupportTicket_List(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	createBody := map[string]interface{}{
		"subject":  "List Test Ticket",
		"category": "technical",
		"priority": "high",
		"message":  "Testing list functionality.",
	}
	ts.doRequest("POST", "/api/v1/support/tickets", createBody, ts.authHeaders(accessToken))

	resp, result := ts.doRequest("GET", "/api/v1/support/tickets", nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_SupportTicket_AddMessage(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	createBody := map[string]interface{}{
		"subject":  "Message Test Ticket",
		"category": "bug_report",
		"priority": "urgent",
		"message":  "Initial message.",
	}
	createResp, createResult := ts.doRequest("POST", "/api/v1/support/tickets", createBody, ts.authHeaders(accessToken))

	if createResp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d: %v", createResp.StatusCode, createResult)
	}

	data, ok := createResult["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data in response: %v", createResult)
	}

	ticketID, ok := data["id"].(string)
	if !ok {
		t.Fatalf("expected ticket id in data: %v", data)
	}

	msgBody := map[string]interface{}{
		"message": "This is a follow-up message.",
	}
	msgResp, msgResult := ts.doRequest("POST", "/api/v1/support/tickets/"+ticketID+"/messages", msgBody, ts.authHeaders(accessToken))

	if msgResp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d: %v", msgResp.StatusCode, msgResult)
	}
}

func TestIntegration_SupportTicket_Get(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	createBody := map[string]interface{}{
		"subject":  "Get Test Ticket",
		"category": "general",
		"priority": "normal",
		"message":  "Testing get functionality.",
	}
	createResp, createResult := ts.doRequest("POST", "/api/v1/support/tickets", createBody, ts.authHeaders(accessToken))

	if createResp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d: %v", createResp.StatusCode, createResult)
	}

	data, ok := createResult["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data in response: %v", createResult)
	}

	ticketID, ok := data["id"].(string)
	if !ok {
		t.Fatalf("expected ticket id in data: %v", data)
	}

	resp, result := ts.doRequest("GET", "/api/v1/support/tickets/"+ticketID, nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}
