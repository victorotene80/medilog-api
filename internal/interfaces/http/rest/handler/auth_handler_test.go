package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/domain/contracts"
	"github.com/victorotene80/medilog-api/test/testutil"
)

func TestAuthHandler_CreateUser_Success(t *testing.T) {
	result := &dto.CreateUserDTO{
		UserID:              123,
		Email:               strPtr("test@example.com"),
		Phone:               nil,
		FirstName:           "John",
		LastName:            "Doe",
		OnboardingCompleted: false,
		RequiresOnboarding:  true,
	}

	bus := newMockBus[command.RegisterCommand, *dto.CreateUserDTO](result, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewAuthHandler(bus, validator)

	body := `{"first_name":"John","last_name":"Doe","email":"test@example.com","password":"password123","dob":"1990-01-01","sex":1,"blood_type":"O+","country_code":"US"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateUser(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestAuthHandler_CreateUser_ValidationError(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(assert.AnError)

	h := NewAuthHandler(bus, validator)

	body := `{"first_name":"J","last_name":"D","password":"123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateUser(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_CreateUser_InvalidJSON(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewAuthHandler(bus, validator)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateUser(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	result := &dto.LoginResultDTO{
		UserID:              "user-123",
		Status:              "active",
		OnboardingCompleted: true,
		RequiresOnboarding:  false,
	}

	bus := newMockBus[command.LoginCommand, *dto.LoginResultDTO](result, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewAuthHandler(bus, validator)

	body := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestAuthHandler_Login_ValidationError(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(assert.AnError)

	h := NewAuthHandler(bus, validator)

	body := `{"email":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// RefreshSessionRequest tags refresh_token as `omitempty`, so the validator
// passes a blank token through; the handler's own guard is what turns it into a
// 400. The permissive mock validator here mirrors that: it must be the guard,
// not validation, producing the response.
func TestAuthHandler_RefreshSession_EmptyToken(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewAuthHandler(bus, validator)

	body := `{"refresh_token":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.RefreshSession(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["status"])
	assert.Equal(t, "Refresh token is required", resp["message"])
}

func TestAuthHandler_DeleteAccount_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewAuthHandler(bus, validator)

	body := `{"otp_code":"123456","recipient":"test@example.com"}`
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/me", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.DeleteAccount(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Logout_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewAuthHandler(bus, validator)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()

	h.Logout(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_ChangePassword_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewAuthHandler(bus, validator)

	body := `{"old_password":"old12345","new_password":"new12345"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ChangePassword(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_ChangePassword_Success(t *testing.T) {
	bus := newMockBus[command.ChangePasswordCommand, struct{}](struct{}{}, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewAuthHandler(bus, validator)

	body := `{"old_password":"old12345","new_password":"new12345"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ChangePassword(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_ForgotPassword_Success(t *testing.T) {
	bus := newMockBus[command.ForgotPasswordCommand, struct{}](struct{}{}, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewAuthHandler(bus, validator)

	body := `{"recipient":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ForgotPassword(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_ResetPassword_Success(t *testing.T) {
	bus := newMockBus[command.ResetPasswordCommand, struct{}](struct{}{}, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewAuthHandler(bus, validator)

	body := `{"recipient":"test@example.com","otp_code":"123456","new_password":"newpass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ResetPassword(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_LoginViaOTP_Success(t *testing.T) {
	result := &contracts.TokenPair{
		AccessToken:  contracts.Token{Value: "access-tok", ExpiresAt: testutil.FutureTime(15 * 60)},
		RefreshToken: contracts.Token{Value: "refresh-tok", ExpiresAt: testutil.FutureTime(7 * 24 * 3600)},
	}

	bus := newMockBus[command.LoginViaOTPCommand, *contracts.TokenPair](result, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewAuthHandler(bus, validator)

	body := `{"recipient":"+15551234567","code":"123456","channel":"sms","device_id":"dev-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login-otp", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.LoginViaOTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_DeleteAccount_OTPRequired(t *testing.T) {
	bus := newMockBus[command.DeleteAccountCommand, struct{}](struct{}{}, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewAuthHandler(bus, validator)

	body := `{"otp_code":"123456","recipient":"test@example.com"}`
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/me", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.DeleteAccount(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func strPtr(s string) *string { return &s }
