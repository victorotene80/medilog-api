package handler

import (
	"net/http"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/httperr"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/mapper"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
)

type DashboardHandler struct {
	commandBus *messaging.CommandBus
	validator  appContracts.Validator
}

func NewDashboardHandler(
	commandBus *messaging.CommandBus,
	validator appContracts.Validator,
) *DashboardHandler {
	return &DashboardHandler{commandBus: commandBus, validator: validator}
}

// GetDashboard godoc
//
//	@Summary     Get dashboard
//	@Description Returns dashboard data for the authenticated user.
//	@Tags        Dashboard
//	@Produce     json
//	@Security    BearerAuth
//	@Success     200 {object} response.APIResponse[response.DashboardResponse]
//	@Failure     401 {object} response.APIResponse[struct{}]
//	@Failure     404 {object} response.APIResponse[struct{}]
//	@Router      /dashboard [get]
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	q := query.GetDashboardQuery{UserID: userID}
	result, err := messaging.Execute[query.GetDashboardQuery, *dto.DashboardDTO](
		h.commandBus,
		r.Context(),
		q,
	)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "DASHBOARD_FETCH_FAILED", "Could not fetch dashboard", err.Error())
		return
	}
	if result == nil {
		response.Error(w, http.StatusNotFound, "DASHBOARD_NOT_FOUND", "Dashboard not found", nil)
		return
	}

	resp := mapper.DashboardDTOToResponse(result)
	response.Success(w, http.StatusOK, "DASHBOARD_FETCHED", "Dashboard retrieved successfully", &resp)
}
