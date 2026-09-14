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

func TestScanHandler_VerifyDrug_Success(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	dName := "Aspirin"
	regNum := "REG-001"
	result := dto.DrugScanDTO{
		PublicID:           "44444444-4444-4444-8444-444444444444",
		DrugName:           &dName,
		RegistrationNumber: &regNum,
		ExpiryDate:         &now,
		IsVerified:         true,
		VerificationStatus: strPtr("verified"),
		CreatedAt:          now,
	}

	bus := newMockBus[command.VerifyDrugScanCommand, dto.DrugScanDTO](result, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewScanHandler(bus, validator)

	body := `{"drug_name":"Aspirin","registration_number":"REG-001","country_code":"NG","expiry_date":"2025-01-15T12:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drugs/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.VerifyDrug(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestScanHandler_VerifyDrug_ValidationError(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(assert.AnError)

	h := NewScanHandler(bus, validator)

	body := `{"drug_name":"Aspirin"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drugs/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.VerifyDrug(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestScanHandler_VerifyDrug_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewScanHandler(bus, validator)

	body := `{"drug_name":"Aspirin","country_code":"NG"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drugs/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.VerifyDrug(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestScanHandler_VerifyDrug_InvalidJSON(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewScanHandler(bus, validator)

	body := `not json`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drugs/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.VerifyDrug(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestScanHandler_ListDrugScans_Success(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	dName := "Aspirin"
	regNum := "REG-001"
	result := []dto.DrugScanDTO{
		{
			PublicID:           "44444444-4444-4444-8444-444444444444",
			DrugName:           &dName,
			RegistrationNumber: &regNum,
			IsVerified:         true,
			CreatedAt:          now,
		},
	}

	bus := newMockBus[query.ListDrugScansQuery, []dto.DrugScanDTO](result, nil)
	validator := new(testutil.MockValidator)

	h := NewScanHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/drugs/scans", nil)
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListDrugScans(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestScanHandler_ListDrugScans_EmptyResult(t *testing.T) {
	result := []dto.DrugScanDTO{}

	bus := newMockBus[query.ListDrugScansQuery, []dto.DrugScanDTO](result, nil)
	validator := new(testutil.MockValidator)

	h := NewScanHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/drugs/scans", nil)
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListDrugScans(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestScanHandler_ListDrugScans_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewScanHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/drugs/scans", nil)
	w := httptest.NewRecorder()

	h.ListDrugScans(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestScanHandler_GetDrugScan_Success(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	dName := "Aspirin"
	regNum := "REG-001"
	result := &dto.DrugScanDTO{
		PublicID:           "44444444-4444-4444-8444-444444444444",
		DrugName:           &dName,
		RegistrationNumber: &regNum,
		ExpiryDate:         &now,
		IsVerified:         true,
		VerificationStatus: strPtr("verified"),
		CreatedAt:          now,
	}

	bus := newMockBus[query.GetDrugScanQuery, *dto.DrugScanDTO](result, nil)
	validator := new(testutil.MockValidator)

	h := NewScanHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "44444444-4444-4444-8444-444444444444")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drugs/scans/44444444-4444-4444-8444-444444444444", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetDrugScan(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestScanHandler_GetDrugScan_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewScanHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "44444444-4444-4444-8444-444444444444")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drugs/scans/44444444-4444-4444-8444-444444444444", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetDrugScan(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestScanHandler_GetDrugScan_MissingID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewScanHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drugs/scans/", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetDrugScan(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestScanHandler_GetDrugScan_NotFound(t *testing.T) {
	bus := newMockBus[query.GetDrugScanQuery, *dto.DrugScanDTO](nil, nil)
	validator := new(testutil.MockValidator)

	h := NewScanHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "33333333-3333-4333-8333-333333333333")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drugs/scans/33333333-3333-4333-8333-333333333333", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetDrugScan(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
