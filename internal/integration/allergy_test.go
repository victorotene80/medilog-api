package integration

import (
	"testing"
)

func TestIntegration_UserAllergy_Add(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	body := map[string]interface{}{
		"allergies": []map[string]interface{}{
			{
				"name":     "Penicillin",
				"category": 1,
				"severity": 3,
			},
		},
	}

	resp, result := ts.doRequest("POST", "/api/v1/health/allergies", body, ts.authHeaders(accessToken))

	if resp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_UserAllergy_List(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	addBody := map[string]interface{}{
		"allergies": []map[string]interface{}{
			{
				"name":     "Aspirin",
				"category": 2,
				"severity": 2,
			},
		},
	}
	ts.doRequest("POST", "/api/v1/health/allergies", addBody, ts.authHeaders(accessToken))

	resp, result := ts.doRequest("GET", "/api/v1/health/allergies", nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_UserAllergy_Delete(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	addBody := map[string]interface{}{
		"allergies": []map[string]interface{}{
			{
				"name":     "Peanuts",
				"category": 1,
				"severity": 5,
			},
		},
	}
	ts.doRequest("POST", "/api/v1/health/allergies", addBody, ts.authHeaders(accessToken))

	listResp, listResult := ts.doRequest("GET", "/api/v1/health/allergies", nil, ts.authHeaders(accessToken))
	if listResp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", listResp.StatusCode)
	}

	listData, ok := listResult["data"].([]interface{})
	if !ok || len(listData) == 0 {
		t.Fatalf("expected allergies in list, got: %v", listResult)
	}

	firstAllergy, ok := listData[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected allergy object, got: %v", listData[0])
	}

	publicID, ok := firstAllergy["id"].(string)
	if !ok {
		t.Fatalf("expected id string, got: %v", firstAllergy)
	}

	resp, result := ts.doRequest("DELETE", "/api/v1/health/allergies/"+publicID, nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_UserAllergy_Unauthenticated(t *testing.T) {
	ts := setupTestServer(t)

	resp, _ := ts.doRequest("GET", "/api/v1/health/allergies", nil, nil)

	if resp.StatusCode != 401 {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}
