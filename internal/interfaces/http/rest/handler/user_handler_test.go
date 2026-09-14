package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
)

type stubValidator struct{}

func (s *stubValidator) Struct(_ any) error { return nil }

func TestGetMe_Success(t *testing.T) {
	now := time.Now()
	email := "john@example.com"
	phone := "+1234567890"
	result := &dto.GetUserDTO{
		ID:        "42",
		Email:     &email,
		Phone:     &phone,
		FirstName: "John",
		LastName:  "Doe",
		Sex:       "male",
		BloodType: "O+",
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	bus := newMockBus[query.GetUserQuery, *dto.GetUserDTO](result, nil)
	h := NewUserHandler(bus, &stubValidator{})

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req = req.WithContext(SetupTestContext("42", "sess-1"))
	rr := httptest.NewRecorder()

	h.GetMe(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var body response.APIResponse[response.GetUserResponse]
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !body.Status {
		t.Fatal("expected status true")
	}
	if body.Data == nil {
		t.Fatal("expected data to be non-nil")
	}
	if body.Data.ID != "42" {
		t.Fatalf("expected user ID 42, got %s", body.Data.ID)
	}
	if body.Data.FirstName != "John" {
		t.Fatalf("expected first name John, got %s", body.Data.FirstName)
	}
}

func TestGetMe_Unauthenticated(t *testing.T) {
	bus := newMockBus[query.GetUserQuery, *dto.GetUserDTO](nil, nil)
	h := NewUserHandler(bus, &stubValidator{})

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rr := httptest.NewRecorder()

	h.GetMe(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}

	var body response.APIResponse[response.EmptyData]
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Status {
		t.Fatal("expected status false")
	}
	if body.Message != "Missing or invalid token" {
		t.Fatalf("unexpected message: %s", body.Message)
	}
}

func TestGetMe_HandlerError(t *testing.T) {
	bus := newMockBus[query.GetUserQuery, *dto.GetUserDTO](nil, errors.New("db timeout"))
	h := NewUserHandler(bus, &stubValidator{})

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req = req.WithContext(SetupTestContext("42", "sess-1"))
	rr := httptest.NewRecorder()

	h.GetMe(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}

	var body response.APIResponse[response.EmptyData]
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Status {
		t.Fatal("expected status false")
	}
}

func TestGetMe_NilResult(t *testing.T) {
	bus := newMockBus[query.GetUserQuery, *dto.GetUserDTO](nil, nil)
	h := NewUserHandler(bus, &stubValidator{})

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req = req.WithContext(SetupTestContext("42", "sess-1"))
	rr := httptest.NewRecorder()

	h.GetMe(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}

	var body response.APIResponse[response.EmptyData]
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Status {
		t.Fatal("expected status false")
	}
	if body.Message != "User not found" {
		t.Fatalf("unexpected message: %s", body.Message)
	}
}

func TestGetMe_InvalidUserID(t *testing.T) {
	bus := newMockBus[query.GetUserQuery, *dto.GetUserDTO](nil, nil)
	h := NewUserHandler(bus, &stubValidator{})

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req = req.WithContext(SetupTestContext("not-a-number", "sess-1"))
	rr := httptest.NewRecorder()

	h.GetMe(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}

	var body response.APIResponse[response.EmptyData]
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Status {
		t.Fatal("expected status false")
	}
}

func TestGetMe_EmptyUserID(t *testing.T) {
	bus := newMockBus[query.GetUserQuery, *dto.GetUserDTO](nil, nil)
	h := NewUserHandler(bus, &stubValidator{})

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req = req.WithContext(SetupTestContext("", "sess-1"))
	rr := httptest.NewRecorder()

	h.GetMe(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}

	var body response.APIResponse[response.EmptyData]
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Status {
		t.Fatal("expected status false")
	}
}
