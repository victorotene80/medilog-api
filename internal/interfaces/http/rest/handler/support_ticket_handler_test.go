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
	"github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/test/testutil"
)

func TestSupportTicketHandler_CreateTicket_Success(t *testing.T) {
	now := testutil.FixedClock()()
	result := &dto.SupportTicketResponseDTO{
		ID:           "ticket-123",
		TicketNumber: "TKT-001",
		Subject:      "Login issue",
		Category:     "technical",
		Status:       "open",
		Priority:     "normal",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	bus := newMockBus[command.CreateSupportTicketCommand, *dto.SupportTicketResponseDTO](result, nil)
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewSupportTicketHandler(bus, validator)

	body := `{"subject":"Login issue","category":"technical","priority":"normal","message":"I cannot login"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/support/tickets", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.CreateTicket(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestSupportTicketHandler_CreateTicket_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(nil)

	h := NewSupportTicketHandler(bus, validator)

	body := `{"subject":"Login issue","category":"technical","priority":"normal","message":"I cannot login"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/support/tickets", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateTicket(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSupportTicketHandler_CreateTicket_ValidationError(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)
	validator.On("Struct", mock.Anything).Return(assert.AnError)

	h := NewSupportTicketHandler(bus, validator)

	body := `{"subject":"Login issue"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/support/tickets", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateTicket(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSupportTicketHandler_ListTickets_Success(t *testing.T) {
	result := &dto.ListSupportTicketsResponseDTO{
		Tickets: []*dto.SupportTicketResponseDTO{},
	}

	bus := newMockBus[command.ListSupportTicketsQuery, *dto.ListSupportTicketsResponseDTO](result, nil)
	validator := new(testutil.MockValidator)

	h := NewSupportTicketHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/support/tickets", nil)
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListTickets(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSupportTicketHandler_ListTickets_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewSupportTicketHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/support/tickets", nil)
	w := httptest.NewRecorder()

	h.ListTickets(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSupportTicketHandler_GetTicket_InvalidID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewSupportTicketHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("ticketId", "invalid")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/support/tickets/invalid", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetTicket(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSupportTicketHandler_GetTicket_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewSupportTicketHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("ticketId", "1")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/support/tickets/1", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetTicket(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSupportTicketHandler_AddMessage_InvalidID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewSupportTicketHandler(bus, validator)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("ticketId", "invalid")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "123", SessionID: "456"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/support/tickets/invalid/messages", bytes.NewBufferString(`{"message":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.AddMessage(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
