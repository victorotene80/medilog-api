package integration

import (
	"testing"
)

func TestIntegration_Feedback_Submit(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	rating := 5
	title := "Great app"
	body := "This is a great application!"
	appVersion := "1.0.0"
	platform := "ios"
	deviceModel := "iPhone 15"

	feedbackBody := map[string]interface{}{
		"rating":       rating,
		"title":        title,
		"message":      body,
		"app_version":  appVersion,
		"platform":     platform,
		"device_model": deviceModel,
	}

	resp, result := ts.doRequest("POST", "/api/v1/feedback", feedbackBody, ts.authHeaders(accessToken))

	if resp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Feedback_SubmitAnonymous(t *testing.T) {
	ts := setupTestServer(t)

	feedbackBody := map[string]interface{}{
		"rating":  4,
		"message": "Anonymous feedback from test.",
	}

	resp, result := ts.doRequest("POST", "/api/v1/feedback", feedbackBody, nil)

	if resp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Feedback_SubmitValidation(t *testing.T) {
	ts := setupTestServer(t)

	feedbackBody := map[string]interface{}{
		"rating": 10,
	}

	resp, _ := ts.doRequest("POST", "/api/v1/feedback", feedbackBody, nil)

	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}
