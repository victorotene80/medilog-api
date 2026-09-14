package integration

import (
	"testing"
)

func TestIntegration_Reference_Countries(t *testing.T) {
	ts := setupTestServer(t)

	resp, result := ts.doRequest("GET", "/api/v1/reference/countries", nil, nil)

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Reference_Allergies(t *testing.T) {
	ts := setupTestServer(t)

	resp, result := ts.doRequest("GET", "/api/v1/reference/allergies", nil, nil)

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Reference_FunFacts(t *testing.T) {
	ts := setupTestServer(t)

	resp, _ := ts.doRequest("GET", "/api/v1/reference/fun-facts", nil, nil)

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestIntegration_Reference_AllergiesMutate_Unauthenticated(t *testing.T) {
	ts := setupTestServer(t)

	body := map[string]interface{}{
		"name":     "Test Allergy",
		"category": 1,
	}

	resp, _ := ts.doRequest("POST", "/api/v1/reference/allergies", body, nil)

	if resp.StatusCode != 401 {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}
