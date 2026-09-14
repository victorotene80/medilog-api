package integration

import (
	"testing"
)

// addAllergy creates one allergy and returns its public id.
func addAllergy(t *testing.T, ts *testServer, token, name string, category int) string {
	t.Helper()

	ts.doRequest("POST", "/api/v1/health/allergies", map[string]interface{}{
		"allergies": []map[string]interface{}{
			{"name": name, "category": category, "severity": 3},
		},
	}, ts.authHeaders(token))

	_, listResult := ts.doRequest("GET", "/api/v1/health/allergies", nil, ts.authHeaders(token))

	data, ok := listResult["data"].([]interface{})
	if !ok {
		t.Fatalf("unexpected allergy list shape: %v", listResult)
	}

	for _, entry := range data {
		m, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		if m["name"] == name {
			id, _ := m["public_id"].(string)
			if id == "" {
				id, _ = m["id"].(string)
			}
			if id == "" {
				t.Fatalf("allergy %q has no public id: %v", name, m)
			}
			return id
		}
	}

	t.Fatalf("allergy %q not found in %v", name, data)
	return ""
}

// Editing used to mean delete-then-re-add, which minted a new public id every
// save. The record must keep its identity across an update.
func TestIntegration_UserAllergy_UpdateKeepsPublicID(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	id := addAllergy(t, ts, token, "Penicillin", 2)

	resp, result := ts.doRequest("PUT", "/api/v1/health/allergies/"+id, map[string]interface{}{
		"name":        "Penicillin V",
		"category":    2,
		"severity":    4,
		"description": "Rash and swelling",
	}, ts.authHeaders(token))

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d: %v", resp.StatusCode, result)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected update response: %v", result)
	}

	gotID, _ := data["public_id"].(string)
	if gotID == "" {
		gotID, _ = data["id"].(string)
	}
	if gotID != id {
		t.Fatalf("update must preserve the public id: was %q, now %q", id, gotID)
	}
	if data["name"] != "Penicillin V" {
		t.Fatalf("name not applied: %v", data)
	}
}

// Renaming onto another of the user's allergies collides with the
// (user_id, lower(name)) uniqueness rule.
func TestIntegration_UserAllergy_UpdateDuplicateNameConflicts(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	addAllergy(t, ts, token, "Peanuts", 1)
	shellfishID := addAllergy(t, ts, token, "Shellfish", 1)

	resp, result := ts.doRequest("PUT", "/api/v1/health/allergies/"+shellfishID,
		map[string]interface{}{"name": "Peanuts", "category": 1},
		ts.authHeaders(token))

	if resp.StatusCode != 409 {
		t.Fatalf("expected 409 renaming onto an existing allergy, got %d: %v",
			resp.StatusCode, result)
	}
}

// Case-insensitively the same name is the same allergy, so renaming to a
// different case is a no-op rather than a conflict with itself.
func TestIntegration_UserAllergy_UpdateSameNameDifferentCaseSucceeds(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	id := addAllergy(t, ts, token, "Latex", 4)

	resp, result := ts.doRequest("PUT", "/api/v1/health/allergies/"+id,
		map[string]interface{}{"name": "latex", "category": 4, "severity": 2},
		ts.authHeaders(token))

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 re-casing a name, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_UserAllergy_UpdateMalformedIDIsBadRequest(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	resp, result := ts.doRequest("PUT", "/api/v1/health/allergies/not-a-uuid",
		map[string]interface{}{"name": "Peanuts", "category": 1}, ts.authHeaders(token))

	if resp.StatusCode != 400 {
		t.Fatalf("expected 400 for a malformed id, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_UserAllergy_UpdateUnknownIDIsNotFound(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	resp, result := ts.doRequest(
		"PUT", "/api/v1/health/allergies/6f1c9f7e-0000-4000-8000-000000000000",
		map[string]interface{}{"name": "Peanuts", "category": 1}, ts.authHeaders(token))

	if resp.StatusCode != 404 {
		t.Fatalf("expected 404 for an unknown id, got %d: %v", resp.StatusCode, result)
	}
}

// One user must never be able to edit another's record.
func TestIntegration_UserAllergy_UpdateOtherUsersAllergyIsNotFound(t *testing.T) {
	ts := setupTestServer(t)

	ownerToken := setupOnboardedUser(t, ts)
	id := addAllergy(t, ts, ownerToken, "Penicillin", 2)

	attackerToken := setupOnboardedUser(t, ts)

	resp, result := ts.doRequest("PUT", "/api/v1/health/allergies/"+id,
		map[string]interface{}{"name": "Hijacked", "category": 1},
		ts.authHeaders(attackerToken))

	if resp.StatusCode != 404 {
		t.Fatalf("another user's allergy must read as missing, got %d: %v",
			resp.StatusCode, result)
	}
}
