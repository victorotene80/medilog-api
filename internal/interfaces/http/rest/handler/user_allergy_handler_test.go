package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/test/testutil"
)

func TestUserAllergyHandler_AddUserAllergies_Success(t *testing.T) {
	bus := newMockBus[command.CreateUserAllergiesCommand, struct{}](struct{}{}, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewUserAllergyHandler(bus, validator)

	body := `{"allergies":[{"name":"Peanuts","category":1,"severity":3,"description":"Severe reaction"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/health/allergies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.AddUserAllergies(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestUserAllergyHandler_AddUserAllergies_ValidationError(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(assert.AnError)

	h := NewUserAllergyHandler(bus, validator)

	body := `{"allergies":[{"category":1}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/health/allergies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.AddUserAllergies(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserAllergyHandler_AddUserAllergies_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewUserAllergyHandler(bus, validator)

	body := `{"allergies":[{"name":"Peanuts","category":1}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/health/allergies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.AddUserAllergies(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserAllergyHandler_AddUserAllergies_InvalidJSON(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewUserAllergyHandler(bus, validator)

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/health/allergies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.AddUserAllergies(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserAllergyHandler_ListUserAllergies_Success(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	result := []dto.UserAllergyDTO{
		{
			PublicID:    "55555555-5555-4555-8555-555555555555",
			Name:        "Peanuts",
			Severity:    int16Ptr(3),
			Category:    1,
			CategoryStr: "Food",
			IsCustom:    false,
			CreatedAt:   now,
		},
	}

	bus := newMockBus[query.ListUserAllergiesQuery, []dto.UserAllergyDTO](result, nil)
	validator := new(testutil.MockValidator)

	h := NewUserAllergyHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/allergies", nil)
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListUserAllergies(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestUserAllergyHandler_ListUserAllergies_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewUserAllergyHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/allergies", nil)
	w := httptest.NewRecorder()

	h.ListUserAllergies(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserAllergyHandler_ListUserAllergies_InvalidCategory(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewUserAllergyHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/allergies?category=99", nil)
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListUserAllergies(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserAllergyHandler_ListUserAllergies_NonNumericCategory(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewUserAllergyHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/allergies?category=abc", nil)
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListUserAllergies(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserAllergyHandler_DeleteUserAllergy_Success(t *testing.T) {
	bus := newMockBus[command.DeleteUserAllergyCommand, struct{}](struct{}{}, nil)
	validator := new(testutil.MockValidator)

	h := NewUserAllergyHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "55555555-5555-4555-8555-555555555555")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/health/allergies/55555555-5555-4555-8555-555555555555", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.DeleteUserAllergy(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestUserAllergyHandler_DeleteUserAllergy_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewUserAllergyHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "55555555-5555-4555-8555-555555555555")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/health/allergies/55555555-5555-4555-8555-555555555555", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.DeleteUserAllergy(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserAllergyHandler_DeleteUserAllergy_MissingID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewUserAllergyHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/health/allergies/", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.DeleteUserAllergy(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func int16Ptr(v int16) *int16 { return &v }
