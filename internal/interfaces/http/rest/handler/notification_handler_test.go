package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
)

// Notifications are addressed by public UUID. A value that is not a UUID is
// rejected with 400 at the handler boundary, before any query runs — the
// public_id column is Postgres `uuid`, so unvalidated text would raise a cast
// error and surface as a 500.
const (
	testNotificationID        = "6f1c9f7e-1111-4111-8111-111111111111"
	testMissingNotificationID = "6f1c9f7e-2222-4222-8222-222222222222"
)

func TestListNotifications_Success(t *testing.T) {
	now := time.Now()
	notifications := []*dto.NotificationResponseDTO{
		{
			ID:        "notif-abc-123",
			UserID:    "1",
			Title:     "Test Notification",
			Body:      "Body text",
			Type:      "info",
			Status:    "unread",
			CreatedAt: now,
		},
	}
	bus := newMockBus[command.ListNotificationsQuery, []*dto.NotificationResponseDTO](notifications, nil)
	h := NewNotificationHandler(bus, nil)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	ctx := context.WithValue(context.Background(), contracts.AuthContextKey, contracts.AuthContext{UserID: "1", SessionID: "session-1"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListNotifications(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestListNotifications_Unauthenticated(t *testing.T) {
	bus := newMockBus[command.ListNotificationsQuery, []*dto.NotificationResponseDTO](nil, nil)
	h := NewNotificationHandler(bus, nil)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	w := httptest.NewRecorder()

	h.ListNotifications(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListNotifications_CommandError(t *testing.T) {
	bus := newMockBus[command.ListNotificationsQuery, []*dto.NotificationResponseDTO](nil, application.NewNotFound("no notifications"))
	h := NewNotificationHandler(bus, nil)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	ctx := context.WithValue(context.Background(), contracts.AuthContextKey, contracts.AuthContext{UserID: "1", SessionID: "session-1"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListNotifications(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListNotifications_EmptyResult(t *testing.T) {
	bus := newMockBus[command.ListNotificationsQuery, []*dto.NotificationResponseDTO]([]*dto.NotificationResponseDTO{}, nil)
	h := NewNotificationHandler(bus, nil)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	ctx := context.WithValue(context.Background(), contracts.AuthContextKey, contracts.AuthContext{UserID: "1", SessionID: "session-1"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListNotifications(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestGetNotification_Success(t *testing.T) {
	now := time.Now()
	notif := &dto.NotificationResponseDTO{
		ID:        "notif-abc-123",
		UserID:    "1",
		Title:     "Test",
		Body:      "Body",
		Type:      "info",
		Status:    "unread",
		CreatedAt: now,
	}
	bus := newMockBus[command.GetNotificationQuery, *dto.NotificationResponseDTO](notif, nil)
	h := NewNotificationHandler(bus, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("notificationId", testNotificationID)
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "1", SessionID: "session-1"})
	req := httptest.NewRequest(http.MethodGet, "/notifications/"+testNotificationID, nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetNotification(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

// Notifications are addressed by public UUID, so only an empty path segment is
// syntactically invalid. A well-formed but unknown id is a 404, not a 400.
func TestGetNotification_EmptyID(t *testing.T) {
	bus := newMockBus[command.GetNotificationQuery, *dto.NotificationResponseDTO](nil, nil)
	h := NewNotificationHandler(bus, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("notificationId", "  ")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "1", SessionID: "session-1"})
	req := httptest.NewRequest(http.MethodGet, "/notifications/", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetNotification(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// An id that is not a UUID is rejected before any query runs. Without this the
// value reaches a Postgres `uuid` column and raises a cast error as a 500.
func TestGetNotification_MalformedID(t *testing.T) {
	bus := newMockBus[command.GetNotificationQuery, *dto.NotificationResponseDTO](nil, nil)
	h := NewNotificationHandler(bus, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("notificationId", "not-a-uuid")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "1", SessionID: "session-1"})
	req := httptest.NewRequest(http.MethodGet, "/notifications/not-a-uuid", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetNotification(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetNotification_Unauthenticated(t *testing.T) {
	bus := newMockBus[command.GetNotificationQuery, *dto.NotificationResponseDTO](nil, nil)
	h := NewNotificationHandler(bus, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("notificationId", testNotificationID)
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest(http.MethodGet, "/notifications/"+testNotificationID, nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetNotification(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetNotification_NotFound(t *testing.T) {
	bus := newMockBus[command.GetNotificationQuery, *dto.NotificationResponseDTO](nil, application.NewNotFound("notification not found"))
	h := NewNotificationHandler(bus, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("notificationId", testMissingNotificationID)
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "1", SessionID: "session-1"})
	req := httptest.NewRequest(http.MethodGet, "/notifications/"+testMissingNotificationID, nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetNotification(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMarkRead_Success(t *testing.T) {
	now := time.Now()
	notif := &dto.NotificationResponseDTO{
		ID:        "notif-abc-123",
		UserID:    "1",
		Title:     "Test",
		Body:      "Body",
		Type:      "info",
		Status:    "read",
		ReadAt:    &now,
		CreatedAt: now,
	}
	bus := newMockBus[command.MarkNotificationReadCommand, *dto.NotificationResponseDTO](notif, nil)
	h := NewNotificationHandler(bus, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("notificationId", testNotificationID)
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "1", SessionID: "session-1"})
	req := httptest.NewRequest(http.MethodPatch, "/notifications/"+testNotificationID+"/read", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.MarkRead(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestMarkRead_EmptyID(t *testing.T) {
	bus := newMockBus[command.MarkNotificationReadCommand, *dto.NotificationResponseDTO](nil, nil)
	h := NewNotificationHandler(bus, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("notificationId", "  ")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "1", SessionID: "session-1"})
	req := httptest.NewRequest(http.MethodPatch, "/notifications//read", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.MarkRead(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMarkRead_MalformedID(t *testing.T) {
	bus := newMockBus[command.MarkNotificationReadCommand, *dto.NotificationResponseDTO](nil, nil)
	h := NewNotificationHandler(bus, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("notificationId", "not-a-uuid")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "1", SessionID: "session-1"})
	req := httptest.NewRequest(http.MethodPatch, "/notifications/not-a-uuid/read", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.MarkRead(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMarkRead_Unauthenticated(t *testing.T) {
	bus := newMockBus[command.MarkNotificationReadCommand, *dto.NotificationResponseDTO](nil, nil)
	h := NewNotificationHandler(bus, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("notificationId", testNotificationID)
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest(http.MethodPatch, "/notifications/"+testNotificationID+"/read", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.MarkRead(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMarkRead_NotFound(t *testing.T) {
	bus := newMockBus[command.MarkNotificationReadCommand, *dto.NotificationResponseDTO](nil, application.NewNotFound("not found"))
	h := NewNotificationHandler(bus, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("notificationId", testMissingNotificationID)
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contracts.AuthContextKey, contracts.AuthContext{UserID: "1", SessionID: "session-1"})
	req := httptest.NewRequest(http.MethodPatch, "/notifications/"+testMissingNotificationID+"/read", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.MarkRead(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMarkAllRead_Success(t *testing.T) {
	bus := newMockBus[command.MarkAllNotificationsReadCommand, struct{}](struct{}{}, nil)
	h := NewNotificationHandler(bus, nil)

	req := httptest.NewRequest(http.MethodPatch, "/notifications/read-all", nil)
	ctx := context.WithValue(context.Background(), contracts.AuthContextKey, contracts.AuthContext{UserID: "1", SessionID: "session-1"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.MarkAllRead(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestMarkAllRead_Unauthenticated(t *testing.T) {
	bus := newMockBus[command.MarkAllNotificationsReadCommand, struct{}](struct{}{}, nil)
	h := NewNotificationHandler(bus, nil)

	req := httptest.NewRequest(http.MethodPatch, "/notifications/read-all", nil)
	w := httptest.NewRecorder()

	h.MarkAllRead(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMarkAllRead_CommandError(t *testing.T) {
	bus := newMockBus[command.MarkAllNotificationsReadCommand, struct{}](struct{}{}, application.NewNotFound("nothing to mark"))
	h := NewNotificationHandler(bus, nil)

	req := httptest.NewRequest(http.MethodPatch, "/notifications/read-all", nil)
	ctx := context.WithValue(context.Background(), contracts.AuthContextKey, contracts.AuthContext{UserID: "1", SessionID: "session-1"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.MarkAllRead(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
