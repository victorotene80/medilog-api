package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/test/testutil"
)

func TestFeedbackHandler_SubmitFeedback_Success(t *testing.T) {
	bus := newMockBus[command.SubmitFeedbackCommand, *dto.FeedbackResponseDTO](&dto.FeedbackResponseDTO{}, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewFeedbackHandler(bus, validator)

	body := `{"rating":5,"title":"Great app","message":"Love it"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/feedback", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.SubmitFeedback(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestFeedbackHandler_SubmitFeedback_ValidationError(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(assert.AnError)

	h := NewFeedbackHandler(bus, validator)

	body := `{"rating":5}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/feedback", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.SubmitFeedback(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFeedbackHandler_GetFeedback_InvalidID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewFeedbackHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("feedbackId", "invalid")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/feedback/invalid", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetFeedback(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFeedbackHandler_SubmitFeedback_WithAuth(t *testing.T) {
	bus := newMockBus[command.SubmitFeedbackCommand, *dto.FeedbackResponseDTO](&dto.FeedbackResponseDTO{}, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewFeedbackHandler(bus, validator)

	body := `{"rating":4,"message":"Good app"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/feedback", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.SubmitFeedback(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}
