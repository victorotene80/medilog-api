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

func TestVisitHandler_CreateVisit_Success(t *testing.T) {
	bus := newMockBus[command.CreateVisitCommand, struct{}](struct{}{}, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewVisitHandler(bus, validator)

	body := `{"visit_date":"2025-01-15T00:00:00Z","hospital_name":"General Hospital"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/visits", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.CreateVisit(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestVisitHandler_CreateVisit_ValidationError(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(assert.AnError)

	h := NewVisitHandler(bus, validator)

	body := `{"hospital_name":"General Hospital"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/visits", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.CreateVisit(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVisitHandler_CreateVisit_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewVisitHandler(bus, validator)

	body := `{"visit_date":"2025-01-15T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/visits", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateVisit(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestVisitHandler_ListVisits_Success(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	result := []dto.VisitDTO{
		{
			ID:        1,
			PublicID:  "66666666-6666-4666-8666-666666666666",
			UserID:    123,
			VisitDate: now,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	bus := newMockBus[query.ListVisitsQuery, []dto.VisitDTO](result, nil)
	validator := new(testutil.MockValidator)

	h := NewVisitHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/visits", nil)
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListVisits(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestVisitHandler_ListVisits_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewVisitHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/visits", nil)
	w := httptest.NewRecorder()

	h.ListVisits(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestVisitHandler_ListVisits_InvalidTimeFormat(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewVisitHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/visits?from=invalid-time", nil)
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListVisits(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVisitHandler_ListVisits_FromWithoutTo(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewVisitHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/visits?from=2025-01-01", nil)
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListVisits(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVisitHandler_GetVisit_Success(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	result := &dto.VisitDTO{
		ID:        1,
		PublicID:  "66666666-6666-4666-8666-666666666666",
		UserID:    123,
		VisitDate: now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	bus := newMockBus[query.GetVisitQuery, *dto.VisitDTO](result, nil)
	validator := new(testutil.MockValidator)

	h := NewVisitHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "66666666-6666-4666-8666-666666666666")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/visits/66666666-6666-4666-8666-666666666666", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetVisit(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestVisitHandler_GetVisit_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewVisitHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/visits/66666666-6666-4666-8666-666666666666", nil)
	w := httptest.NewRecorder()

	h.GetVisit(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestVisitHandler_GetVisit_MissingID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewVisitHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/visits/", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetVisit(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVisitHandler_UpdateVisit_Success(t *testing.T) {
	bus := newMockBus[command.UpdateVisitCommand, struct{}](struct{}{}, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewVisitHandler(bus, validator)

	body := `{"visit_date":"2025-01-15T00:00:00Z","hospital_name":"Updated Hospital"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/visits/66666666-6666-4666-8666-666666666666", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "66666666-6666-4666-8666-666666666666")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.UpdateVisit(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestVisitHandler_UpdateVisit_ValidationError(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(assert.AnError)

	h := NewVisitHandler(bus, validator)

	body := `{"hospital_name":"Updated Hospital"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/visits/66666666-6666-4666-8666-666666666666", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "66666666-6666-4666-8666-666666666666")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.UpdateVisit(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVisitHandler_UpdateVisit_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewVisitHandler(bus, validator)

	body := `{"visit_date":"2025-01-15T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/visits/66666666-6666-4666-8666-666666666666", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateVisit(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestVisitHandler_UpdateVisit_MissingID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewVisitHandler(bus, validator)

	body := `{"visit_date":"2025-01-15T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/visits/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.UpdateVisit(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVisitHandler_DeleteVisit_Success(t *testing.T) {
	bus := newMockBus[command.DeleteVisitCommand, struct{}](struct{}{}, nil)
	validator := new(testutil.MockValidator)

	h := NewVisitHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "66666666-6666-4666-8666-666666666666")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/visits/66666666-6666-4666-8666-666666666666", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.DeleteVisit(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestVisitHandler_DeleteVisit_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewVisitHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "66666666-6666-4666-8666-666666666666")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/visits/66666666-6666-4666-8666-666666666666", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.DeleteVisit(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestVisitHandler_DeleteVisit_MissingID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewVisitHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/visits/", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.DeleteVisit(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
