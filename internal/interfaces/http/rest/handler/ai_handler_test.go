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
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/test/testutil"
)

func TestAIHandler_CreateConversation_Success(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	result := dto.AIConversationDTO{
		ID:        1,
		PublicID:  "11111111-1111-4111-8111-111111111111",
		UserID:    123,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	bus := newMockBus[command.CreateAIConversationCommand, dto.AIConversationDTO](result, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewAIHandler(bus, validator)

	body := `{"title":"Medication questions"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/conversations", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.CreateConversation(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestAIHandler_CreateConversation_ValidationError(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(assert.AnError)

	h := NewAIHandler(bus, validator)

	body := `{"title":"Medication questions"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/conversations", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.CreateConversation(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAIHandler_CreateConversation_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewAIHandler(bus, validator)

	body := `{"title":"Medication questions"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/conversations", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateConversation(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAIHandler_ListConversations_Success(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	result := []dto.AIConversationDTO{
		{
			ID:        1,
			PublicID:  "11111111-1111-4111-8111-111111111111",
			UserID:    123,
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	bus := newMockBus[query.ListAIConversationsQuery, []dto.AIConversationDTO](result, nil)
	validator := new(testutil.MockValidator)

	h := NewAIHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/conversations", nil)
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListConversations(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestAIHandler_ListConversations_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewAIHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/conversations", nil)
	w := httptest.NewRecorder()

	h.ListConversations(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAIHandler_GetConversation_Success(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	result := &dto.AIConversationDTO{
		ID:        1,
		PublicID:  "11111111-1111-4111-8111-111111111111",
		UserID:    123,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	bus := newMockBus[query.GetAIConversationQuery, *dto.AIConversationDTO](result, nil)
	validator := new(testutil.MockValidator)

	h := NewAIHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "11111111-1111-4111-8111-111111111111")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/conversations/11111111-1111-4111-8111-111111111111", nil)
	ctx := setupTestContextWithRoute("123", "456", rctx)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetConversation(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestAIHandler_GetConversation_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewAIHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/conversations/11111111-1111-4111-8111-111111111111", nil)
	w := httptest.NewRecorder()

	h.GetConversation(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAIHandler_GetConversation_EmptyID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewAIHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/conversations/", nil)
	ctx := setupTestContextWithRoute("123", "456", rctx)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetConversation(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAIHandler_SendMessage_Success(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	result := &dto.SendAIMessageResultDTO{
		Conversation: dto.AIConversationDTO{
			ID:        1,
			PublicID:  "11111111-1111-4111-8111-111111111111",
			UserID:    123,
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Message: dto.AIMessageDTO{
			ID:        1,
			Role:      "user",
			Content:   "What are the side effects?",
			CreatedAt: now,
		},
		Reply: dto.AIMessageDTO{
			ID:        2,
			Role:      "assistant",
			Content:   "Common side effects include...",
			CreatedAt: now,
		},
		Model:        "gpt-4",
		PromptTokens: 100,
		OutputTokens: 50,
	}

	bus := newMockBus[command.SendAIMessageCommand, *dto.SendAIMessageResultDTO](result, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewAIHandler(bus, validator)

	body := `{"message":"What are the side effects?","language":"en"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/conversations/11111111-1111-4111-8111-111111111111/messages", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "11111111-1111-4111-8111-111111111111")
	ctx := setupTestContextWithRoute("123", "456", rctx)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.SendMessage(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestAIHandler_SendMessage_ValidationError(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(assert.AnError)

	h := NewAIHandler(bus, validator)

	body := `{"message":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/conversations/11111111-1111-4111-8111-111111111111/messages", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "11111111-1111-4111-8111-111111111111")
	ctx := setupTestContextWithRoute("123", "456", rctx)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.SendMessage(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAIHandler_SendMessage_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewAIHandler(bus, validator)

	body := `{"message":"What are the side effects?"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/conversations/11111111-1111-4111-8111-111111111111/messages", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.SendMessage(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAIHandler_SendMessage_EmptyID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewAIHandler(bus, validator)

	body := `{"message":"What are the side effects?"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/conversations//messages", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "")
	ctx := setupTestContextWithRoute("123", "456", rctx)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.SendMessage(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAIHandler_ArchiveConversation_Success(t *testing.T) {
	bus := newMockBus[command.ArchiveAIConversationCommand, struct{}](struct{}{}, nil)
	validator := new(testutil.MockValidator)

	h := NewAIHandler(bus, validator)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/ai/conversations/11111111-1111-4111-8111-111111111111/archive", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "11111111-1111-4111-8111-111111111111")
	ctx := setupTestContextWithRoute("123", "456", rctx)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ArchiveConversation(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestAIHandler_ArchiveConversation_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewAIHandler(bus, validator)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/ai/conversations/11111111-1111-4111-8111-111111111111/archive", nil)
	w := httptest.NewRecorder()

	h.ArchiveConversation(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAIHandler_ArchiveConversation_EmptyID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewAIHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("publicId", "")
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/ai/conversations//archive", nil)
	ctx := setupTestContextWithRoute("123", "456", rctx)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ArchiveConversation(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func setupTestContextWithRoute(userID, sessionID string, rctx *chi.Context) context.Context {
	ctx := SetupTestContext(userID, sessionID)
	return context.WithValue(ctx, chi.RouteCtxKey, rctx)
}
