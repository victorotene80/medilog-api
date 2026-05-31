package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/httperr"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/mapper"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/request"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
)

type VisitHandler struct {
	commandBus *messaging.CommandBus
	validator  appContracts.Validator
}

func NewVisitHandler(
	commandBus *messaging.CommandBus,
	validator appContracts.Validator,
) *VisitHandler {
	return &VisitHandler{commandBus: commandBus, validator: validator}
}

// CreateVisit godoc
//
//	@Summary     Add a visit
//	@Tags        Visits
//	@Accept      json
//	@Produce     json
//	@Security    BearerAuth
//	@Param       body body request.CreateVisitRequest true "Visit payload"
//	@Success     201 {object} response.APIResponse[struct{}]
//	@Router      /visits [post]
func (h *VisitHandler) CreateVisit(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	req, ok := decodeAndValidate[request.CreateVisitRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.CreateVisitCommand{
		UserID:         userID,
		HospitalName:   req.HospitalName,
		Diagnosis:      req.Diagnosis,
		VisitDate:      req.VisitDate,
		Outcome:        req.Outcome,
		MedsCount:      intValue(req.MedsCount),
		Doctor:         req.Doctor,
		ChiefComplaint: req.ChiefComplaint,
		Notes:          req.Notes,
		BloodPressure:  req.BloodPressure,
		Temperature:    req.Temperature,
		Weight:         req.Weight,
		Pulse:          req.Pulse,
	}

	_, err := messaging.Execute[command.CreateVisitCommand, struct{}](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "VISIT_CREATE_FAILED", "Could not create visit", err.Error())
		return
	}

	writeEmptySuccess(w, http.StatusCreated, "VISIT_CREATED", "Visit added successfully")
}

// ListVisits godoc
//
//	@Summary     List visits
//	@Tags        Visits
//	@Produce     json
//	@Security    BearerAuth
//	@Param       from query string false "Start datetime or date"
//	@Param       to   query string false "End datetime or date"
//	@Success     200 {object} response.APIResponse[[]response.VisitResponse]
//	@Router      /visits [get]
func (h *VisitHandler) ListVisits(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	from, ok := parseOptionalTimeQuery(w, r, "from")
	if !ok {
		return
	}
	to, ok := parseOptionalTimeQuery(w, r, "to")
	if !ok {
		return
	}
	if (from == nil) != (to == nil) {
		response.Error(w, http.StatusBadRequest, "INVALID_DATE_RANGE", "from and to must be provided together", nil)
		return
	}

	q := query.ListVisitsQuery{UserID: userID, From: from, To: to}

	result, err := messaging.Execute[query.ListVisitsQuery, []dto.VisitDTO](
		h.commandBus, r.Context(), q,
	)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "VISITS_FETCH_FAILED", "Could not fetch visits", err.Error())
		return
	}

	resp := mapper.VisitDTOsToResponse(result)
	response.Success[[]response.VisitResponse](w, http.StatusOK, "VISITS_FETCHED", "Visits retrieved successfully", &resp)
}

// GetVisit godoc
//
//	@Summary     Get a visit
//	@Tags        Visits
//	@Produce     json
//	@Security    BearerAuth
//	@Param       publicId path string true "Visit public ID"
//	@Success     200 {object} response.APIResponse[response.VisitResponse]
//	@Router      /visits/{publicId} [get]
func (h *VisitHandler) GetVisit(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	publicID := chi.URLParam(r, "publicId")
	if publicID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Visit public ID is required", nil)
		return
	}

	q := query.GetVisitQuery{UserID: userID, PublicID: publicID}

	result, err := messaging.Execute[query.GetVisitQuery, *dto.VisitDTO](
		h.commandBus, r.Context(), q,
	)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "VISIT_FETCH_FAILED", "Could not fetch visit", err.Error())
		return
	}
	if result == nil {
		response.Error(w, http.StatusNotFound, "VISIT_NOT_FOUND", "Visit not found", nil)
		return
	}

	resp := mapper.VisitDTOToResponse(*result)
	response.Success[response.VisitResponse](w, http.StatusOK, "VISIT_FETCHED", "Visit retrieved successfully", &resp)
}

// UpdateVisit godoc
//
//	@Summary     Update a visit
//	@Tags        Visits
//	@Accept      json
//	@Produce     json
//	@Security    BearerAuth
//	@Param       publicId path string                     true "Visit public ID"
//	@Param       body     body request.UpdateVisitRequest true "Updated visit"
//	@Success     200 {object} response.APIResponse[struct{}]
//	@Router      /visits/{publicId} [put]
func (h *VisitHandler) UpdateVisit(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	publicID := chi.URLParam(r, "publicId")
	if publicID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Visit public ID is required", nil)
		return
	}

	req, ok := decodeAndValidate[request.UpdateVisitRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.UpdateVisitCommand{
		UserID:         userID,
		PublicID:       publicID,
		HospitalName:   req.HospitalName,
		Diagnosis:      req.Diagnosis,
		VisitDate:      req.VisitDate,
		Outcome:        req.Outcome,
		MedsCount:      intValue(req.MedsCount),
		Doctor:         req.Doctor,
		ChiefComplaint: req.ChiefComplaint,
		Notes:          req.Notes,
		BloodPressure:  req.BloodPressure,
		Temperature:    req.Temperature,
		Weight:         req.Weight,
		Pulse:          req.Pulse,
	}

	_, err := messaging.Execute[command.UpdateVisitCommand, struct{}](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "VISIT_UPDATE_FAILED", "Could not update visit", err.Error())
		return
	}

	writeEmptySuccess(w, http.StatusOK, "VISIT_UPDATED", "Visit updated successfully")
}

// DeleteVisit godoc
//
//	@Summary     Delete a visit
//	@Tags        Visits
//	@Produce     json
//	@Security    BearerAuth
//	@Param       publicId path string true "Visit public ID"
//	@Success     200 {object} response.APIResponse[struct{}]
//	@Router      /visits/{publicId} [delete]
func (h *VisitHandler) DeleteVisit(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	publicID := chi.URLParam(r, "publicId")
	if publicID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Visit public ID is required", nil)
		return
	}

	cmd := command.DeleteVisitCommand{UserID: userID, PublicID: publicID}

	_, err := messaging.Execute[command.DeleteVisitCommand, struct{}](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "VISIT_DELETE_FAILED", "Could not delete visit", err.Error())
		return
	}

	writeEmptySuccess(w, http.StatusOK, "VISIT_DELETED", "Visit deleted successfully")
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func parseOptionalTimeQuery(w http.ResponseWriter, r *http.Request, key string) (*time.Time, bool) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, true
	}

	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return &parsed, true
	}

	parsedDate, err := time.Parse("2006-01-02", raw)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_DATE", key+" must be RFC3339 datetime or YYYY-MM-DD date", nil)
		return nil, false
	}
	return &parsedDate, true
}
