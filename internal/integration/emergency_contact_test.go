package integration

import (
	"fmt"
	"testing"
)

// contactBody builds a valid create/update payload. Phones must be E.164.
func contactBody(name, phone string, isPrimary bool) map[string]interface{} {
	return map[string]interface{}{
		"name":         name,
		"relationship": "Sister",
		"phone":        phone,
		"country_code": "NG",
		"is_primary":   isPrimary,
	}
}

// The regression test for the onboarding lockout.
//
// POST /emergency-contacts is the only endpoint that sets
// is_onboarding_completed, but it used to sit behind middleware that required
// onboarding to already be complete — so a verified user who had not yet added
// a contact was permanently stuck on a 403 pointing at this very endpoint.
// Before the fix this test fails with 403 ONBOARDING_REQUIRED.
func TestIntegration_EmergencyContact_ReachableBeforeOnboarding(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	_, registerResult := ts.registerUser(email, "Password123!", "Ada", "Lovelace")
	activateUserWithoutOnboarding(t, ts.db, ts.getUserID(registerResult))

	_, loginResult := ts.loginUser(email, "Password123!")
	token := ts.getAccessToken(loginResult)

	resp, result := ts.doRequest(
		"POST", "/api/v1/emergency-contacts",
		contactBody("Jane Doe", randomTestPhone(), false),
		ts.authHeaders(token),
	)

	if resp.StatusCode != 201 {
		t.Fatalf("a not-yet-onboarded user must be able to add their first contact; got %d: %v",
			resp.StatusCode, result)
	}

	// Creating the contact is what completes onboarding.
	meResp, meResult := ts.doRequest("GET", "/api/v1/users/me", nil, ts.authHeaders(token))
	if meResp.StatusCode != 200 {
		t.Fatalf("expected 200 from /users/me, got %d: %v", meResp.StatusCode, meResult)
	}

	data, ok := meResult["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected /users/me shape: %v", meResult)
	}

	if completed, _ := data["onboarding_completed"].(bool); !completed {
		t.Fatalf("expected onboarding_completed=true after adding a contact, got: %v", data)
	}
}

// The first contact must be primary even when the client does not ask for it,
// otherwise the user ends up with no primary at all.
func TestIntegration_EmergencyContact_FirstContactIsPrimary(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	_, result := ts.doRequest(
		"POST", "/api/v1/emergency-contacts",
		contactBody("Jane Doe", randomTestPhone(), false),
		ts.authHeaders(token),
	)

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected create response: %v", result)
	}

	if isPrimary, _ := data["is_primary"].(bool); !isPrimary {
		t.Fatalf("the first contact must be primary, got: %v", data)
	}
}

// is_primary used to be dropped when building the command, so every contact
// was created non-primary regardless of what the client sent.
func TestIntegration_EmergencyContact_IsPrimaryIsHonoured(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	ts.doRequest("POST", "/api/v1/emergency-contacts",
		contactBody("First", randomTestPhone(), false), ts.authHeaders(token))

	_, second := ts.doRequest("POST", "/api/v1/emergency-contacts",
		contactBody("Second", randomTestPhone(), true), ts.authHeaders(token))

	data, ok := second["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected create response: %v", second)
	}

	if isPrimary, _ := data["is_primary"].(bool); !isPrimary {
		t.Fatalf("is_primary=true must be honoured, got: %v", data)
	}

	// Exactly one primary must survive.
	_, listResult := ts.doRequest("GET", "/api/v1/emergency-contacts", nil, ts.authHeaders(token))
	contacts := extractContacts(t, listResult)

	primaries := 0
	for _, c := range contacts {
		if p, _ := c["is_primary"].(bool); p {
			primaries++
		}
	}

	if primaries != 1 {
		t.Fatalf("expected exactly one primary contact, got %d: %v", primaries, contacts)
	}
}

func TestIntegration_EmergencyContact_ListUpdateDelete(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	_, first := ts.doRequest("POST", "/api/v1/emergency-contacts",
		contactBody("First", randomTestPhone(), true), ts.authHeaders(token))
	firstID := contactID(t, first)

	ts.doRequest("POST", "/api/v1/emergency-contacts",
		contactBody("Second", randomTestPhone(), false), ts.authHeaders(token))

	_, listResult := ts.doRequest("GET", "/api/v1/emergency-contacts", nil, ts.authHeaders(token))
	if got := len(extractContacts(t, listResult)); got != 2 {
		t.Fatalf("expected 2 contacts, got %d: %v", got, listResult)
	}

	// Update keeps the same id — the whole point of having a PUT.
	updateResp, updateResult := ts.doRequest(
		"PUT", "/api/v1/emergency-contacts/"+firstID,
		contactBody("First Renamed", randomTestPhone(), true),
		ts.authHeaders(token),
	)
	if updateResp.StatusCode != 200 {
		t.Fatalf("expected 200 from update, got %d: %v", updateResp.StatusCode, updateResult)
	}

	updated, _ := updateResult["data"].(map[string]interface{})
	if updated["id"] != firstID {
		t.Fatalf("update must preserve the public id: was %v, now %v", firstID, updated["id"])
	}
	if updated["name"] != "First Renamed" {
		t.Fatalf("expected the new name, got %v", updated["name"])
	}

	delResp, delResult := ts.doRequest(
		"DELETE", "/api/v1/emergency-contacts/"+firstID, nil, ts.authHeaders(token))
	if delResp.StatusCode != 200 {
		t.Fatalf("expected 200 from delete, got %d: %v", delResp.StatusCode, delResult)
	}

	_, afterList := ts.doRequest("GET", "/api/v1/emergency-contacts", nil, ts.authHeaders(token))
	if got := len(extractContacts(t, afterList)); got != 1 {
		t.Fatalf("expected 1 contact after delete, got %d: %v", got, afterList)
	}
}

// Deleting the last contact would silently un-onboard the user.
func TestIntegration_EmergencyContact_CannotDeleteOnlyContact(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	_, created := ts.doRequest("POST", "/api/v1/emergency-contacts",
		contactBody("Only", randomTestPhone(), true), ts.authHeaders(token))

	resp, result := ts.doRequest(
		"DELETE", "/api/v1/emergency-contacts/"+contactID(t, created), nil, ts.authHeaders(token))

	if resp.StatusCode != 409 {
		t.Fatalf("expected 409 deleting the only contact, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_EmergencyContact_DuplicatePhoneConflicts(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	phone := randomTestPhone()
	ts.doRequest("POST", "/api/v1/emergency-contacts",
		contactBody("First", phone, true), ts.authHeaders(token))

	resp, result := ts.doRequest("POST", "/api/v1/emergency-contacts",
		contactBody("Duplicate", phone, false), ts.authHeaders(token))

	if resp.StatusCode != 409 {
		t.Fatalf("expected 409 for a duplicate phone, got %d: %v", resp.StatusCode, result)
	}
}

// A malformed public id must be a 400, not a 500. These columns are Postgres
// uuid, so passing arbitrary text into the query raises a cast error.
func TestIntegration_EmergencyContact_MalformedIDIsBadRequest(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	resp, result := ts.doRequest(
		"DELETE", "/api/v1/emergency-contacts/not-a-uuid", nil, ts.authHeaders(token))

	if resp.StatusCode != 400 {
		t.Fatalf("expected 400 for a malformed id, got %d: %v", resp.StatusCode, result)
	}
}

// A well-formed but unknown id is a 404.
func TestIntegration_EmergencyContact_UnknownIDIsNotFound(t *testing.T) {
	ts := setupTestServer(t)
	token := setupOnboardedUser(t, ts)

	resp, result := ts.doRequest(
		"DELETE", "/api/v1/emergency-contacts/6f1c9f7e-0000-4000-8000-000000000000",
		nil, ts.authHeaders(token))

	if resp.StatusCode != 404 {
		t.Fatalf("expected 404 for an unknown id, got %d: %v", resp.StatusCode, result)
	}
}

// setupOnboardedUser registers, onboards and logs in a user, returning the token.
func setupOnboardedUser(t *testing.T, ts *testServer) string {
	t.Helper()

	email := randomTestEmail()
	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")

	return ts.getAccessToken(loginResult)
}

func extractContacts(t *testing.T, result map[string]interface{}) []map[string]interface{} {
	t.Helper()

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected list response: %v", result)
	}

	raw, ok := data["contacts"].([]interface{})
	if !ok {
		t.Fatalf("expected data.contacts array, got: %v", data)
	}

	contacts := make([]map[string]interface{}, 0, len(raw))
	for _, c := range raw {
		m, ok := c.(map[string]interface{})
		if !ok {
			t.Fatalf("unexpected contact entry: %v", c)
		}
		contacts = append(contacts, m)
	}

	return contacts
}

func contactID(t *testing.T, result map[string]interface{}) string {
	t.Helper()

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected create response: %v", result)
	}

	id, ok := data["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected a contact id, got: %v", data)
	}

	return fmt.Sprint(id)
}
