package integration

import (
	"testing"
)

func TestIntegration_UpdateMe_RoundTrip(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	body := map[string]interface{}{
		"first_name":       "Ada",
		"last_name":        "Lovelace",
		"sex":              2,
		"blood_type":       "O+",
		"country_code":     "NG",
		"height":           170.5,
		"weight":           64.2,
		"weight_unit":      "kg",
		"temperature_unit": "celsius",
	}

	resp, result := ts.doRequest("PATCH", "/api/v1/users/me", body, ts.authHeaders(token))
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d: %v", resp.StatusCode, result)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected response shape: %v", result)
	}

	if data["first_name"] != "Ada" || data["last_name"] != "Lovelace" {
		t.Fatalf("name not applied: %v", data)
	}
	if data["blood_type"] != "O+" {
		t.Fatalf("blood_type not applied: %v", data)
	}
	if data["height"] != 170.5 || data["weight"] != 64.2 {
		t.Fatalf("height/weight not applied: %v", data)
	}

	// The change must survive a re-read, not just be echoed back.
	_, meResult := ts.doRequest("GET", "/api/v1/users/me", nil, ts.authHeaders(token))
	meData, _ := meResult["data"].(map[string]interface{})
	if meData["first_name"] != "Ada" || meData["height"] != 170.5 {
		t.Fatalf("update did not persist: %v", meData)
	}
}

// Omitted fields must be left alone — this is a PATCH, not a replace.
func TestIntegration_UpdateMe_PartialLeavesOtherFields(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	ts.doRequest("PATCH", "/api/v1/users/me",
		map[string]interface{}{"first_name": "Ada", "blood_type": "O+"},
		ts.authHeaders(token))

	ts.doRequest("PATCH", "/api/v1/users/me",
		map[string]interface{}{"last_name": "Lovelace"},
		ts.authHeaders(token))

	_, meResult := ts.doRequest("GET", "/api/v1/users/me", nil, ts.authHeaders(token))
	data, _ := meResult["data"].(map[string]interface{})

	if data["first_name"] != "Ada" {
		t.Fatalf("first_name was clobbered by a later partial update: %v", data)
	}
	if data["blood_type"] != "O+" {
		t.Fatalf("blood_type was clobbered by a later partial update: %v", data)
	}
	if data["last_name"] != "Lovelace" {
		t.Fatalf("last_name not applied: %v", data)
	}
}

// Email and phone are not writable here; unknown fields are rejected outright
// so the exclusion is loud rather than silent.
func TestIntegration_UpdateMe_RejectsEmailAndPhone(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	for _, field := range []string{"email", "phone"} {
		resp, result := ts.doRequest("PATCH", "/api/v1/users/me",
			map[string]interface{}{field: "attacker@example.com"},
			ts.authHeaders(token))

		if resp.StatusCode != 400 {
			t.Fatalf("expected 400 when %q is present, got %d: %v", field, resp.StatusCode, result)
		}
	}
}

func TestIntegration_UpdateMe_RejectsInvalidBloodType(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	resp, result := ts.doRequest("PATCH", "/api/v1/users/me",
		map[string]interface{}{"blood_type": "Z+"}, ts.authHeaders(token))

	if resp.StatusCode != 400 && resp.StatusCode != 422 {
		t.Fatalf("expected a validation failure for a bogus blood type, got %d: %v",
			resp.StatusCode, result)
	}
}

func TestIntegration_NotificationPreferences_RoundTrip(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	resp, result := ts.doRequest("GET", "/api/v1/users/me/notification-preferences",
		nil, ts.authHeaders(token))
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d: %v", resp.StatusCode, result)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected response shape: %v", result)
	}

	// Defaults from NewDefaultUserProfile.
	if enabled, _ := data["medication_reminders_enabled"].(bool); !enabled {
		t.Fatalf("expected medication reminders on by default: %v", data)
	}
	if data["timezone"] != "UTC" {
		t.Fatalf("expected UTC default timezone, got %v", data["timezone"])
	}

	// This is the bug the frontend reported: toggles looked functional but were
	// only held in memory, so every choice was lost on reload.
	patchResp, patchResult := ts.doRequest("PATCH", "/api/v1/users/me/notification-preferences",
		map[string]interface{}{
			"medication_reminders_enabled": false,
			"push_enabled":                 false,
			"timezone":                     "Africa/Lagos",
		}, ts.authHeaders(token))
	if patchResp.StatusCode != 200 {
		t.Fatalf("expected 200 from patch, got %d: %v", patchResp.StatusCode, patchResult)
	}

	_, afterResult := ts.doRequest("GET", "/api/v1/users/me/notification-preferences",
		nil, ts.authHeaders(token))
	after, _ := afterResult["data"].(map[string]interface{})

	if enabled, _ := after["medication_reminders_enabled"].(bool); enabled {
		t.Fatalf("medication_reminders_enabled did not persist as false: %v", after)
	}
	if enabled, _ := after["push_enabled"].(bool); enabled {
		t.Fatalf("push_enabled did not persist as false: %v", after)
	}
	if after["timezone"] != "Africa/Lagos" {
		t.Fatalf("timezone did not persist: %v", after["timezone"])
	}
	// Untouched toggles keep their value.
	if enabled, _ := after["appointment_reminders_enabled"].(bool); !enabled {
		t.Fatalf("an untouched toggle was changed: %v", after)
	}
}

func TestIntegration_NotificationPreferences_RejectsBadTimezone(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	resp, result := ts.doRequest("PATCH", "/api/v1/users/me/notification-preferences",
		map[string]interface{}{"timezone": "Mars/Olympus_Mons"}, ts.authHeaders(token))

	if resp.StatusCode != 422 {
		t.Fatalf("expected 422 for an unresolvable timezone, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_AIQuota_Shape(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	resp, result := ts.doRequest("GET", "/api/v1/ai/quota", nil, ts.authHeaders(token))
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d: %v", resp.StatusCode, result)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected response shape: %v", result)
	}

	// The frontend's quota UI was unreachable because this reported 0/0.
	total, _ := data["total"].(float64)
	if total <= 0 {
		t.Fatalf("expected a non-zero question allowance, got %v", data)
	}

	used, _ := data["used"].(float64)
	if used != 0 {
		t.Fatalf("a fresh user should have used nothing, got %v", data)
	}

	remaining, _ := data["remaining"].(float64)
	if remaining != total-used {
		t.Fatalf("remaining should be total-used, got %v", data)
	}

	if _, ok := data["resets_at"].(string); !ok {
		t.Fatalf("expected a resets_at timestamp: %v", data)
	}

	if isPro, _ := data["is_pro"].(bool); isPro {
		t.Fatalf("a fresh user should not be pro: %v", data)
	}
}

func TestIntegration_AIQuota_Unauthenticated(t *testing.T) {
	ts := setupTestServer(t)

	resp, _ := ts.doRequest("GET", "/api/v1/ai/quota", nil, nil)
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}
