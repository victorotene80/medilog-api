package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/test/testutil"
)

func TestDashboardHandler_GetDashboard_Success(t *testing.T) {
	result := &dto.DashboardDTO{
		User: dto.DashboardUserDTO{
			ID:        "123",
			FirstName: "Jane",
			LastName:  "Doe",
		},
		Medications: []dto.DashboardMedicationDTO{},
		HealthOverview: dto.DashboardHealthOverviewDTO{
			Visits: dto.DashboardVisitOverviewDTO{
				Total:    5,
				Upcoming: 1,
			},
			Medications: dto.DashboardMedicationOverviewDTO{
				Completed: 3,
				Total:     5,
			},
			FlaggedDrugs: dto.DashboardFlaggedDrugOverviewDTO{
				Flagged: 0,
				Total:   0,
			},
		},
	}

	bus := newMockBus[query.GetDashboardQuery, *dto.DashboardDTO](result, nil)
	validator := new(testutil.MockValidator)

	h := NewDashboardHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetDashboard(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["status"])
}

func TestDashboardHandler_GetDashboard_Unauthenticated(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewDashboardHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	w := httptest.NewRecorder()

	h.GetDashboard(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["status"])
}

func TestDashboardHandler_GetDashboard_HandlerError(t *testing.T) {
	bus := newMockBus[query.GetDashboardQuery, *dto.DashboardDTO](nil, errors.New("db timeout"))
	validator := new(testutil.MockValidator)

	h := NewDashboardHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetDashboard(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["status"])
}

func TestDashboardHandler_GetDashboard_NilResult(t *testing.T) {
	bus := newMockBus[query.GetDashboardQuery, *dto.DashboardDTO](nil, nil)
	validator := new(testutil.MockValidator)

	h := NewDashboardHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetDashboard(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["status"])
}

func TestDashboardHandler_GetDashboard_InvalidUserID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewDashboardHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	ctx := SetupTestContext("not-a-number", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetDashboard(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["status"])
}

func TestDashboardHandler_GetDashboard_EmptyUserID(t *testing.T) {
	bus := messaging.NewCommandBus()
	validator := new(testutil.MockValidator)

	h := NewDashboardHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	ctx := SetupTestContext("", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetDashboard(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["status"])
}
