package handler

import (
	"net/http"
	"strconv"

	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/httperr"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/mapper"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
)

type AuditLogHandler struct {
	commandBus *messaging.CommandBus
	validator  contracts.Validator
}

func NewAuditLogHandler(
	commandBus *messaging.CommandBus,
	validator contracts.Validator,
) *AuditLogHandler {
	return &AuditLogHandler{
		commandBus: commandBus,
		validator:  validator,
	}
}

func (h *AuditLogHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	// No isAdmin flag: it was derived from "is authenticated", which is a
	// different question, and the route is already gated by AdminMiddleware.
	cmd := command.ListAuditLogsQuery{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	result, err := messaging.Execute[
		command.ListAuditLogsQuery,
		*dto.ListAuditLogsDTO,
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(
			w,
			httperr.StatusFrom(err),
			"LIST_AUDIT_LOGS_FAILED",
			"Could not list audit logs",
			err,
		)
		return
	}

	resp := mapper.ListAuditLogsDTOToResponse(result)

	response.Success[response.ListAuditLogsResponse](
		w,
		http.StatusOK,
		"AUDIT_LOGS_FETCHED",
		"Audit logs retrieved successfully",
		&resp,
	)
}
