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

func TestMedicationHandler_CreateMedication_Success(t *testing.T) {
	bus := newMockBus[command.CreateMedicationCommand, struct{}](struct{}{}, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewMedicationHandler(bus, validator)

	body := `{"name":"Aspirin","dosage":"100mg","frequency":"daily","with_food":true,"times":[{"time_value":"08:00"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medications", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.CreateMedication(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestMedicationHandler_CreateMedication_ValidationError(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(assert.AnError)

	h := NewMedicationHandler(bus, validator)

	body := `{"dosage":"100mg"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medications", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.CreateMedication(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMedicationHandler_CreateMedication_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewMedicationHandler(bus, validator)

	body := `{"name":"Aspirin","dosage":"100mg","frequency":"daily"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medications", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateMedication(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMedicationHandler_ListMedications_Success(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	result := []dto.MedicationDTO{
		{
			ID:        1,
			PublicID:  "22222222-2222-4222-8222-222222222222",
			UserID:    123,
			Name:      "Aspirin",
			Dosage:    strPtr("100mg"),
			Times:     []dto.MedicationTimeDTO{{ID: 1, TimeValue: "08:00"}},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	bus := newMockBus[query.ListMedicationsQuery, []dto.MedicationDTO](result, nil)
	validator := new(testutil.MockValidator)

	h := NewMedicationHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/medications", nil)
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListMedications(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestMedicationHandler_ListMedications_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewMedicationHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/medications", nil)
	w := httptest.NewRecorder()

	h.ListMedications(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMedicationHandler_GetMedication_Success(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	result := &dto.MedicationDTO{
		ID:        1,
		PublicID:  "22222222-2222-4222-8222-222222222222",
		UserID:    123,
		Name:      "Aspirin",
		Dosage:    strPtr("100mg"),
		Times:     []dto.MedicationTimeDTO{{ID: 1, TimeValue: "08:00"}},
		CreatedAt: now,
		UpdatedAt: now,
	}

	bus := newMockBus[query.GetMedicationQuery, *dto.MedicationDTO](result, nil)
	validator := new(testutil.MockValidator)

	h := NewMedicationHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "22222222-2222-4222-8222-222222222222")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/medications/22222222-2222-4222-8222-222222222222", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetMedication(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestMedicationHandler_GetMedication_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewMedicationHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/medications/22222222-2222-4222-8222-222222222222", nil)
	w := httptest.NewRecorder()

	h.GetMedication(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMedicationHandler_GetMedication_MissingID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewMedicationHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/medications/", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetMedication(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMedicationHandler_UpdateMedication_Success(t *testing.T) {
	bus := newMockBus[command.UpdateMedicationCommand, struct{}](struct{}{}, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewMedicationHandler(bus, validator)

	body := `{"name":"Ibuprofen","dosage":"200mg","frequency":"twice daily"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/medications/22222222-2222-4222-8222-222222222222", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "22222222-2222-4222-8222-222222222222")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.UpdateMedication(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestMedicationHandler_UpdateMedication_ValidationError(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(assert.AnError)

	h := NewMedicationHandler(bus, validator)

	body := `{"dosage":"200mg"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/medications/22222222-2222-4222-8222-222222222222", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.UpdateMedication(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMedicationHandler_UpdateMedication_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewMedicationHandler(bus, validator)

	body := `{"name":"Ibuprofen","dosage":"200mg"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/medications/22222222-2222-4222-8222-222222222222", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateMedication(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMedicationHandler_UpdateMedication_MissingID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewMedicationHandler(bus, validator)

	body := `{"name":"Ibuprofen","dosage":"200mg"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/medications/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.UpdateMedication(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMedicationHandler_CompleteMedication_Success(t *testing.T) {
	bus := newMockBus[command.CompleteMedicationCommand, struct{}](struct{}{}, nil)
	validator := new(testutil.MockValidator)

	h := NewMedicationHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "22222222-2222-4222-8222-222222222222")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/medications/22222222-2222-4222-8222-222222222222/complete", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.CompleteMedication(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestMedicationHandler_CompleteMedication_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewMedicationHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "22222222-2222-4222-8222-222222222222")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/medications/22222222-2222-4222-8222-222222222222/complete", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.CompleteMedication(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMedicationHandler_CompleteMedication_MissingID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewMedicationHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/medications//complete", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.CompleteMedication(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMedicationHandler_DeleteMedication_Success(t *testing.T) {
	bus := newMockBus[command.DeleteMedicationCommand, struct{}](struct{}{}, nil)
	validator := new(testutil.MockValidator)

	h := NewMedicationHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "22222222-2222-4222-8222-222222222222")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/medications/22222222-2222-4222-8222-222222222222", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.DeleteMedication(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestMedicationHandler_DeleteMedication_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewMedicationHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "22222222-2222-4222-8222-222222222222")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/medications/22222222-2222-4222-8222-222222222222", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.DeleteMedication(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMedicationHandler_DeleteMedication_MissingID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewMedicationHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/medications/", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.DeleteMedication(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMedicationHandler_LogAdherence_Success(t *testing.T) {
	medResult := &dto.MedicationDTO{
		ID:       1,
		PublicID: "22222222-2222-4222-8222-222222222222",
		UserID:   123,
		Name:     "Aspirin",
	}

	bus := newMockBus[query.GetMedicationQuery, *dto.MedicationDTO](medResult, nil)
	messaging.MustRegister[command.LogMedicationAdherenceCommand, struct{}](bus, &mockCommandHandler[command.LogMedicationAdherenceCommand, struct{}]{result: struct{}{}, err: nil})
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewMedicationHandler(bus, validator)

	body := `{"scheduled_at":"2025-01-15T08:00:00Z","status":2,"note":"Took after breakfast"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medications/22222222-2222-4222-8222-222222222222/adherence", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "22222222-2222-4222-8222-222222222222")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.LogAdherence(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestMedicationHandler_LogAdherence_ValidationError(t *testing.T) {
	medResult := &dto.MedicationDTO{
		ID:       1,
		PublicID: "22222222-2222-4222-8222-222222222222",
		UserID:   123,
		Name:     "Aspirin",
	}

	bus := newMockBus[query.GetMedicationQuery, *dto.MedicationDTO](medResult, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(assert.AnError)

	h := NewMedicationHandler(bus, validator)

	body := `{"status":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medications/22222222-2222-4222-8222-222222222222/adherence", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "22222222-2222-4222-8222-222222222222")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.LogAdherence(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMedicationHandler_LogAdherence_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewMedicationHandler(bus, validator)

	body := `{"scheduled_at":"2025-01-15T08:00:00Z","status":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medications/22222222-2222-4222-8222-222222222222/adherence", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.LogAdherence(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMedicationHandler_LogAdherence_MissingID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewMedicationHandler(bus, validator)

	body := `{"scheduled_at":"2025-01-15T08:00:00Z","status":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medications//adherence", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.LogAdherence(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
