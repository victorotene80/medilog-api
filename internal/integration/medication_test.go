package integration

import (
	"testing"
)

func TestIntegration_Medication_Create(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	body := map[string]interface{}{
		"name":      "Paracetamol",
		"dosage":    "500mg",
		"frequency": "daily",
		"with_food": false,
		"times":     []map[string]interface{}{{"time_value": "08:00"}},
	}

	resp, result := ts.doRequest("POST", "/api/v1/medications", body, ts.authHeaders(accessToken))

	if resp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Medication_CreateValidation(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	body := map[string]interface{}{}

	resp, _ := ts.doRequest("POST", "/api/v1/medications", body, ts.authHeaders(accessToken))

	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestIntegration_Medication_List(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	createBody := map[string]interface{}{
		"name":      "Lisinopril",
		"dosage":    "10mg",
		"frequency": "daily",
		"times":     []map[string]interface{}{{"time_value": "08:00"}},
	}
	ts.doRequest("POST", "/api/v1/medications", createBody, ts.authHeaders(accessToken))

	resp, result := ts.doRequest("GET", "/api/v1/medications", nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Medication_Get(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	createBody := map[string]interface{}{
		"name":      "Amoxicillin",
		"dosage":    "250mg",
		"frequency": "three_times_daily",
		"times":     []map[string]interface{}{{"time_value": "08:00"}, {"time_value": "14:00"}, {"time_value": "20:00"}},
	}
	ts.doRequest("POST", "/api/v1/medications", createBody, ts.authHeaders(accessToken))

	listResp, listResult := ts.doRequest("GET", "/api/v1/medications", nil, ts.authHeaders(accessToken))
	if listResp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", listResp.StatusCode)
	}

	listData, ok := listResult["data"].([]interface{})
	if !ok || len(listData) == 0 {
		t.Fatalf("expected medications in list, got: %v", listResult)
	}

	firstMed, ok := listData[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected medication object, got: %v", listData[0])
	}

	publicID, ok := firstMed["public_id"].(string)
	if !ok {
		t.Fatalf("expected public_id string, got: %v", firstMed)
	}

	resp, result := ts.doRequest("GET", "/api/v1/medications/"+publicID, nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Medication_GetNotFound(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	resp, _ := ts.doRequest("GET", "/api/v1/medications/00000000-0000-0000-0000-000000000000", nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 404 {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestIntegration_Medication_Update(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	createBody := map[string]interface{}{
		"name":      "Metformin",
		"dosage":    "500mg",
		"frequency": "daily",
		"times":     []map[string]interface{}{{"time_value": "08:00"}},
	}
	ts.doRequest("POST", "/api/v1/medications", createBody, ts.authHeaders(accessToken))

	listResp, listResult := ts.doRequest("GET", "/api/v1/medications", nil, ts.authHeaders(accessToken))
	if listResp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", listResp.StatusCode)
	}

	listData, ok := listResult["data"].([]interface{})
	if !ok || len(listData) == 0 {
		t.Fatalf("expected medications in list, got: %v", listResult)
	}

	firstMed, ok := listData[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected medication object, got: %v", listData[0])
	}

	publicID, ok := firstMed["public_id"].(string)
	if !ok {
		t.Fatalf("expected public_id string, got: %v", firstMed)
	}

	updateBody := map[string]interface{}{
		"name":      "Metformin XR",
		"dosage":    "1000mg",
		"frequency": "twice_daily",
		"times":     []map[string]interface{}{{"time_value": "08:00"}},
	}
	resp, result := ts.doRequest("PUT", "/api/v1/medications/"+publicID, updateBody, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Medication_Complete(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	createBody := map[string]interface{}{
		"name":      "Ciprofloxacin",
		"dosage":    "500mg",
		"frequency": "twice_daily",
		"times":     []map[string]interface{}{{"time_value": "08:00"}, {"time_value": "20:00"}},
	}
	ts.doRequest("POST", "/api/v1/medications", createBody, ts.authHeaders(accessToken))

	listResp, listResult := ts.doRequest("GET", "/api/v1/medications", nil, ts.authHeaders(accessToken))
	if listResp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", listResp.StatusCode)
	}

	listData, ok := listResult["data"].([]interface{})
	if !ok || len(listData) == 0 {
		t.Fatalf("expected medications in list, got: %v", listResult)
	}

	firstMed, ok := listData[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected medication object, got: %v", listData[0])
	}

	publicID, ok := firstMed["public_id"].(string)
	if !ok {
		t.Fatalf("expected public_id string, got: %v", firstMed)
	}

	resp, result := ts.doRequest("PATCH", "/api/v1/medications/"+publicID+"/complete", nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Medication_Delete(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	createBody := map[string]interface{}{
		"name":      "Omeprazole",
		"dosage":    "20mg",
		"frequency": "daily",
		"times":     []map[string]interface{}{{"time_value": "08:00"}},
	}
	ts.doRequest("POST", "/api/v1/medications", createBody, ts.authHeaders(accessToken))

	listResp, listResult := ts.doRequest("GET", "/api/v1/medications", nil, ts.authHeaders(accessToken))
	if listResp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", listResp.StatusCode)
	}

	listData, ok := listResult["data"].([]interface{})
	if !ok || len(listData) == 0 {
		t.Fatalf("expected medications in list, got: %v", listResult)
	}

	firstMed, ok := listData[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected medication object, got: %v", listData[0])
	}

	publicID, ok := firstMed["public_id"].(string)
	if !ok {
		t.Fatalf("expected public_id string, got: %v", firstMed)
	}

	resp, result := ts.doRequest("DELETE", "/api/v1/medications/"+publicID, nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Medication_Unauthenticated(t *testing.T) {
	ts := setupTestServer(t)

	resp, _ := ts.doRequest("GET", "/api/v1/medications", nil, nil)

	if resp.StatusCode != 401 {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}
