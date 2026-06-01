package handler

import (
	"net/http"

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

type FunFactHandler struct {
	commandBus *messaging.CommandBus
	validator  appContracts.Validator
}

func NewFunFactHandler(
	commandBus *messaging.CommandBus,
	validator appContracts.Validator,
) *FunFactHandler {
	return &FunFactHandler{commandBus: commandBus, validator: validator}
}

// ListFunFacts godoc
//
//	@Summary     List fun facts
//	@Description Returns fun facts, optionally limited to active facts.
//	@Tags        Reference
//	@Produce     json
//	@Param       active_only query bool false "Return only active fun facts"
//	@Success     200 {object} response.APIResponse[[]response.FunFactResponse]
//	@Failure     500 {object} response.APIResponse[struct{}]
//	@Router      /reference/fun-facts/ [get]
func (h *FunFactHandler) ListFunFacts(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") == "true"
	q := query.ListFunFactsQuery{ActiveOnly: activeOnly}

	result, err := messaging.Execute[query.ListFunFactsQuery, []dto.FunFactDTO](
		h.commandBus,
		r.Context(),
		q,
	)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "FUN_FACTS_FETCH_FAILED", "Could not fetch fun facts", err.Error())
		return
	}

	resp := mapper.FunFactDTOsToResponse(result)
	response.Success(w, http.StatusOK, "FUN_FACTS_FETCHED", "Fun facts retrieved successfully", &resp)
}

// GetFunFact godoc
//
//	@Summary     Get fun fact
//	@Description Returns one fun fact by numeric ID.
//	@Tags        Reference
//	@Produce     json
//	@Param       id path int true "Fun fact ID"
//	@Success     200 {object} response.APIResponse[response.FunFactResponse]
//	@Failure     400 {object} response.APIResponse[struct{}]
//	@Failure     404 {object} response.APIResponse[struct{}]
//	@Router      /reference/fun-facts/{id} [get]
func (h *FunFactHandler) GetFunFact(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Fun fact ID must be a positive integer", nil)
		return
	}

	q := query.GetFunFactQuery{ID: id}
	result, err := messaging.Execute[query.GetFunFactQuery, *dto.FunFactDTO](h.commandBus, r.Context(), q)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "FUN_FACT_FETCH_FAILED", "Could not fetch fun fact", err.Error())
		return
	}
	if result == nil {
		response.Error(w, http.StatusNotFound, "FUN_FACT_NOT_FOUND", "Fun fact not found", nil)
		return
	}

	resp := mapper.FunFactDTOToResponse(*result)
	response.Success(w, http.StatusOK, "FUN_FACT_FETCHED", "Fun fact retrieved successfully", &resp)
}

// CreateFunFact godoc
//
//	@Summary     Create fun fact
//	@Description Creates a fun fact reference record.
//	@Tags        Reference
//	@Accept      json
//	@Produce     json
//	@Security    BearerAuth
//	@Param       body body request.CreateFunFactRequest true "Fun fact payload"
//	@Success     201 {object} response.APIResponse[struct{}]
//	@Failure     400 {object} response.APIResponse[struct{}]
//	@Failure     401 {object} response.APIResponse[struct{}]
//	@Router      /reference/fun-facts/ [post]
func (h *FunFactHandler) CreateFunFact(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[request.CreateFunFactRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.CreateFunFactCommand{
		Title:             req.Title,
		Text:              req.Text,
		Category:          req.Category,
		TargetCountryCode: req.TargetCountryCode,
		TargetAgeMin:      req.TargetAgeMin,
		TargetAgeMax:      req.TargetAgeMax,
		AllergyCategory:   req.AllergyCategory,
		IsActive:          req.IsActive,
	}

	_, err := messaging.Execute[command.CreateFunFactCommand, struct{}](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "FUN_FACT_CREATE_FAILED", "Could not create fun fact", err.Error())
		return
	}

	writeEmptySuccess(w, http.StatusCreated, "FUN_FACT_CREATED", "Fun fact created successfully")
}

// UpdateFunFact godoc
//
//	@Summary     Update fun fact
//	@Description Updates a fun fact reference record by numeric ID.
//	@Tags        Reference
//	@Accept      json
//	@Produce     json
//	@Security    BearerAuth
//	@Param       id   path int                          true "Fun fact ID"
//	@Param       body body request.UpdateFunFactRequest true "Updated fun fact payload"
//	@Success     200 {object} response.APIResponse[struct{}]
//	@Failure     400 {object} response.APIResponse[struct{}]
//	@Failure     401 {object} response.APIResponse[struct{}]
//	@Failure     404 {object} response.APIResponse[struct{}]
//	@Router      /reference/fun-facts/{id} [put]
func (h *FunFactHandler) UpdateFunFact(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Fun fact ID must be a positive integer", nil)
		return
	}

	req, ok := decodeAndValidate[request.UpdateFunFactRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.UpdateFunFactCommand{
		ID:                id,
		Title:             req.Title,
		Text:              req.Text,
		Category:          req.Category,
		TargetCountryCode: req.TargetCountryCode,
		TargetAgeMin:      req.TargetAgeMin,
		TargetAgeMax:      req.TargetAgeMax,
		AllergyCategory:   req.AllergyCategory,
		IsActive:          req.IsActive,
	}

	_, err = messaging.Execute[command.UpdateFunFactCommand, struct{}](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "FUN_FACT_UPDATE_FAILED", "Could not update fun fact", err.Error())
		return
	}

	writeEmptySuccess(w, http.StatusOK, "FUN_FACT_UPDATED", "Fun fact updated successfully")
}

// DeleteFunFact godoc
//
//	@Summary     Delete fun fact
//	@Description Deletes a fun fact reference record by numeric ID.
//	@Tags        Reference
//	@Produce     json
//	@Security    BearerAuth
//	@Param       id path int true "Fun fact ID"
//	@Success     200 {object} response.APIResponse[struct{}]
//	@Failure     400 {object} response.APIResponse[struct{}]
//	@Failure     401 {object} response.APIResponse[struct{}]
//	@Failure     404 {object} response.APIResponse[struct{}]
//	@Router      /reference/fun-facts/{id} [delete]
func (h *FunFactHandler) DeleteFunFact(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Fun fact ID must be a positive integer", nil)
		return
	}

	cmd := command.DeleteFunFactCommand{ID: id}
	_, err = messaging.Execute[command.DeleteFunFactCommand, struct{}](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "FUN_FACT_DELETE_FAILED", "Could not delete fun fact", err.Error())
		return
	}

	writeEmptySuccess(w, http.StatusOK, "FUN_FACT_DELETED", "Fun fact deleted successfully")
}
