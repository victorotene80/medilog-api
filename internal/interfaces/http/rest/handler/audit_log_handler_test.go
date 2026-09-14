package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/test/testutil"
)

func TestAuditLogHandler_ListAuditLogs_Success(t *testing.T) {
	result := &dto.ListAuditLogsDTO{
		Logs:  []*dto.AuditLogDTO{},
		Total: 0,
	}
	bus := newMockBus[command.ListAuditLogsQuery, *dto.ListAuditLogsDTO](result, nil)
	validator := new(testutil.MockValidator)

	h := NewAuditLogHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs", nil)
	ctx := SetupTestContext("123", "456")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListAuditLogs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditLogHandler_ListAuditLogs_WithParams(t *testing.T) {
	result := &dto.ListAuditLogsDTO{
		Logs:  []*dto.AuditLogDTO{},
		Total: 0,
	}
	bus := newMockBus[command.ListAuditLogsQuery, *dto.ListAuditLogsDTO](result, nil)
	validator := new(testutil.MockValidator)

	h := NewAuditLogHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs?user_id=123&limit=10&offset=20", nil)
	ctx := SetupTestContext("456", "789")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListAuditLogs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditLogHandler_ListAuditLogs_DefaultParams(t *testing.T) {
	result := &dto.ListAuditLogsDTO{
		Logs:  []*dto.AuditLogDTO{},
		Total: 0,
	}
	bus := newMockBus[command.ListAuditLogsQuery, *dto.ListAuditLogsDTO](result, nil)
	validator := new(testutil.MockValidator)

	h := NewAuditLogHandler(bus, validator)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs?limit=abc&offset=-5", nil)
	ctx := SetupTestContext("456", "789")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListAuditLogs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
