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
	"github.com/victorotene80/medilog-api/test/testutil"
)

func TestEmergencyContactHandler_CreateEmergencyContact_Success(t *testing.T) {
	bus := newMockBus[command.EmergencyContactCommand, *dto.EmergencyContactDTO](&dto.EmergencyContactDTO{}, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewEmergencyContactHandler(bus, validator)

	body := `{"name":"Jane Doe","relationship":"Mother","phone":"+1234567890","country_code":"US","is_primary":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/emergency-contacts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.CreateEmergencyContact(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestEmergencyContactHandler_CreateEmergencyContact_ValidationError(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(assert.AnError)

	h := NewEmergencyContactHandler(bus, validator)

	body := `{"name":"Jane"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/emergency-contacts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.CreateEmergencyContact(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmergencyContactHandler_CreateEmergencyContact_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewEmergencyContactHandler(bus, validator)

	body := `{"name":"Jane Doe","relationship":"Mother","phone":"+1234567890","country_code":"US"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/emergency-contacts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateEmergencyContact(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestEmergencyContactHandler_CreateEmergencyContact_InvalidJSON(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewEmergencyContactHandler(bus, validator)

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/emergency-contacts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.CreateEmergencyContact(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmergencyContactHandler_CreateEmergencyContact_UnknownField(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewEmergencyContactHandler(bus, validator)

	body := `{"name":"Jane","relationship":"Mother","phone":"+1234567890","country_code":"US","unknown_field":"bad"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/emergency-contacts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.CreateEmergencyContact(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
