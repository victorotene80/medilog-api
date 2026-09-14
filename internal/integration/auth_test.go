package integration

import (
	"fmt"
	"testing"
)

func TestIntegration_Register_Success(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	resp, result := ts.registerUser(email, "Password123!", "John", "Doe")

	if resp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Register_DuplicateEmail(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerUser(email, "Password123!", "John", "Doe")

	resp, result := ts.registerUser(email, "Password123!", "Jane", "Doe")

	if resp.StatusCode != 409 {
		t.Fatalf("expected status 409, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Register_ValidationError(t *testing.T) {
	ts := setupTestServer(t)

	resp, result := ts.registerUser("not-an-email", "123", "", "")

	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Login_Success(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerUser(email, "Password123!", "John", "Doe")

	resp, result := ts.loginUser(email, "Password123!")

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}

	accessToken := ts.getAccessToken(result)
	if accessToken == "" {
		t.Fatal("expected non-empty access token")
	}

	refreshToken := ts.getRefreshToken(result)
	if refreshToken == "" {
		t.Fatal("expected non-empty refresh token")
	}
}

func TestIntegration_Login_InvalidCredentials(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerUser(email, "Password123!", "John", "Doe")

	resp, result := ts.loginUser(email, "WrongPassword!")

	if resp.StatusCode != 401 {
		t.Fatalf("expected status 401, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Login_NonExistentUser(t *testing.T) {
	ts := setupTestServer(t)

	resp, result := ts.loginUser("nonexistent@example.com", "Password123!")

	if resp.StatusCode != 401 {
		t.Fatalf("expected status 401, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_RefreshToken_Success(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerUser(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")

	refreshToken := ts.getRefreshToken(loginResult)

	body := map[string]interface{}{
		"refresh_token": refreshToken,
	}

	resp, result := ts.doRequest("POST", "/api/v1/auth/refresh", body, nil)

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}

	newAccessToken := ts.getAccessToken(result)
	if newAccessToken == "" {
		t.Fatal("expected non-empty new access token")
	}
}

func TestIntegration_RefreshToken_InvalidToken(t *testing.T) {
	ts := setupTestServer(t)

	body := map[string]interface{}{
		"refresh_token": "invalid-token-12345",
	}

	resp, result := ts.doRequest("POST", "/api/v1/auth/refresh", body, nil)

	if resp.StatusCode != 401 {
		t.Fatalf("expected status 401, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_ChangePassword_Success(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerUser(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	body := map[string]interface{}{
		"old_password": "Password123!",
		"new_password": "NewPassword456!",
	}

	resp, result := ts.doRequest("POST", "/api/v1/auth/change-password", body, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %v", resp.StatusCode, result)
	}

	ts.loginUser(email, "NewPassword456!")
}

func TestIntegration_ChangePassword_WrongOldPassword(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerUser(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	body := map[string]interface{}{
		"old_password": "WrongPassword!",
		"new_password": "NewPassword456!",
	}

	resp, _ := ts.doRequest("POST", "/api/v1/auth/change-password", body, ts.authHeaders(accessToken))

	if resp.StatusCode != 400 && resp.StatusCode != 401 {
		t.Fatalf("expected status 400 or 401, got %d", resp.StatusCode)
	}
}

func TestIntegration_DeleteAccount_Unauthenticated(t *testing.T) {
	ts := setupTestServer(t)

	resp, _ := ts.doRequest("DELETE", "/api/v1/users/me", nil, nil)

	if resp.StatusCode != 401 {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestIntegration_DeleteAccount_OTPRequired(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	// DELETE /users/me sits behind RequireOnboardingCompleted, so a freshly
	// registered user is turned away with 403 VERIFICATION_REQUIRED before the
	// handler ever runs — which says nothing about whether deletion demands an
	// OTP. The user has to be active and onboarded for the 400 to mean anything.
	ts.registerAndCompleteOnboarding(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	resp, result := ts.doRequest("DELETE", "/api/v1/users/me", nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d: %v", resp.StatusCode, result)
	}
}

func TestIntegration_Logout_Success(t *testing.T) {
	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerUser(email, "Password123!", "John", "Doe")
	_, loginResult := ts.loginUser(email, "Password123!")
	accessToken := ts.getAccessToken(loginResult)

	resp, _ := ts.doRequest("POST", "/api/v1/auth/logout", nil, ts.authHeaders(accessToken))

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	fmt.Println("Integration tests passed")
}

func TestIntegration_OTP_Request(t *testing.T) {
	ts := setupTestServer(t)

	body := map[string]interface{}{
		"recipient": "+2348012345678",
		"channel":   "sms",
		"purpose":   "login",
	}

	resp, _ := ts.doRequest("POST", "/api/v1/auth/otp/request", body, nil)

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestIntegration_OTP_VerifyInvalidCode(t *testing.T) {
	ts := setupTestServer(t)

	body := map[string]interface{}{
		"recipient": "+2348012345678",
		"code":      "000000",
		"channel":   "sms",
		"purpose":   "login",
	}

	resp, _ := ts.doRequest("POST", "/api/v1/auth/otp/verify", body, nil)

	if resp.StatusCode != 400 && resp.StatusCode != 401 && resp.StatusCode != 422 {
		t.Fatalf("expected status 400, 401, or 422, got %d", resp.StatusCode)
	}
}

func TestIntegration_ForgotPassword_Success(t *testing.T) {
	ts := setupTestServer(t)

	body := map[string]interface{}{
		"recipient": "test@example.com",
	}

	resp, _ := ts.doRequest("POST", "/api/v1/auth/forgot-password", body, nil)

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

// The password reset flow had no integration coverage, and a hardcoded
// recipient in OTPCodeRepository.FindLatestByRecipientAndPurpose survived in
// the working tree because of it: the query ignored its recipient argument and
// always looked up one specific test address. Every reset either failed
// outright or, where that address had a live OTP, reset the wrong account's
// password.
//
// These two tests pin the contract that broke: a reset must find the OTP
// belonging to the recipient it was asked about, and must change that
// account's password and no other.
//
// APP_ENV is set per-test rather than in setupTestServer's defaults because
// IsLive also switches off rate limiting, which other tests here rely on. In a
// non-live environment ForgotPassword stores the hash of a known code, so the
// reset below exercises the real OTP lookup and verification rather than the
// "123456" bypass.
func TestIntegration_ResetPassword_UpdatesRequestedAccount(t *testing.T) {
	t.Setenv("APP_ENV", "test")

	ts := setupTestServer(t)
	email := randomTestEmail()

	ts.registerUser(email, "Password123!", "John", "Doe")

	resp, result := ts.doRequest("POST", "/api/v1/auth/forgot-password",
		map[string]interface{}{"recipient": email}, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("forgot-password: expected 200, got %d: %v", resp.StatusCode, result)
	}

	resp, result = ts.doRequest("POST", "/api/v1/auth/reset-password", map[string]interface{}{
		"recipient":    email,
		"otp_code":     "123456",
		"new_password": "NewPassword456!",
	}, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("reset-password: expected 200, got %d: %v", resp.StatusCode, result)
	}

	resp, result = ts.loginUser(email, "NewPassword456!")
	if resp.StatusCode != 200 {
		t.Fatalf("login with the new password: expected 200, got %d: %v", resp.StatusCode, result)
	}

	resp, _ = ts.loginUser(email, "Password123!")
	if resp.StatusCode == 200 {
		t.Fatal("login with the old password succeeded after a reset")
	}
}

func TestIntegration_ResetPassword_LeavesOtherAccountsAlone(t *testing.T) {
	t.Setenv("APP_ENV", "test")

	ts := setupTestServer(t)
	target, bystander := randomTestEmail(), randomTestEmail()

	// Asserted because a silent duplicate here would leave the test operating on
	// one account and failing much later, in the assertion below.
	for name, email := range map[string]string{"target": target, "bystander": bystander} {
		resp, result := ts.registerUser(email, "Password123!", "Some", "User")
		if resp.StatusCode != 201 {
			t.Fatalf("registering the %s: expected 201, got %d: %v", name, resp.StatusCode, result)
		}
	}

	// Both accounts have an outstanding reset OTP, so a lookup that ignores the
	// recipient can pick up the wrong one.
	for _, email := range []string{target, bystander} {
		resp, result := ts.doRequest("POST", "/api/v1/auth/forgot-password",
			map[string]interface{}{"recipient": email}, nil)
		if resp.StatusCode != 200 {
			t.Fatalf("forgot-password for %s: expected 200, got %d: %v", email, resp.StatusCode, result)
		}
	}

	resp, result := ts.doRequest("POST", "/api/v1/auth/reset-password", map[string]interface{}{
		"recipient":    target,
		"otp_code":     "123456",
		"new_password": "NewPassword456!",
	}, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("reset-password: expected 200, got %d: %v", resp.StatusCode, result)
	}

	resp, result = ts.loginUser(bystander, "Password123!")
	if resp.StatusCode != 200 {
		t.Fatalf("the bystander's password was changed by a reset aimed at another account: got %d: %v",
			resp.StatusCode, result)
	}
}
