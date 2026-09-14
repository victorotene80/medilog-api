package handler

import (
	"net/http"
	"strconv"

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

type AllergyHandler struct {
	commandBus *messaging.CommandBus
	validator  appContracts.Validator
}

func NewAllergyHandler(
	commandBus *messaging.CommandBus,
	validator appContracts.Validator,
) *AllergyHandler {
	return &AllergyHandler{
		commandBus: commandBus,
		validator:  validator,
	}
}

// ListAllergies godoc
//
//	@Summary     List master allergies
//	@Description Returns the full reference list. Optionally filter by ?category=1..5
//	@Tags        Reference
//	@Produce     json
//	@Param       category query int false "Allergy category (1-5)"
//	@Success     200 {object} response.APIResponse[[]response.AllergyResponse]
//	@Router      /reference/allergies/ [get]
func (h *AllergyHandler) ListAllergies(w http.ResponseWriter, r *http.Request) {
	q := query.GetAllergiesQuery{}

	if raw := r.URL.Query().Get("category"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 5 {
			response.Error(w, http.StatusBadRequest, "INVALID_CATEGORY", "category must be an integer between 1 and 5", nil)
			return
		}
		q.Category = &v
	}

	result, err := messaging.Execute[query.GetAllergiesQuery, []dto.AllergyDTO](
		h.commandBus, r.Context(), q,
	)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err), "ALLERGIES_FETCH_FAILED", "Could not fetch allergies", err)
		return
	}

	resp := mapper.AllergyDTOsToResponse(result)
	response.Success(w, http.StatusOK, "ALLERGIES_FETCHED", "Allergies retrieved successfully", &resp)
}

// AddAllergy godoc
//
//	@Summary     Add a master allergy
//	@Description Adds a new entry to the reference allergy list (admin use).
//	@Tags        Reference
//	@Accept      json
//	@Produce     json
//	@Param       body body request.AddAllergyRequest true "Allergy payload"
//	@Success     201 {object} response.APIResponse[response.EmptyData]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Router      /reference/allergies/ [post]
func (h *AllergyHandler) AddAllergy(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[request.AddAllergyRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.AllergyCommand{
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
	}

	_, err := messaging.Execute[command.AllergyCommand, struct{}](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err), "ALLERGY_CREATE_FAILED", "Could not create allergy", err)
		return
	}

	writeEmptySuccess(w, http.StatusCreated, "ALLERGY_CREATED", "Allergy created successfully")
}

// UpdateAllergy godoc
//
//	@Summary     Update a master allergy
//	@Description Updates name, category, and description of an existing reference allergy.
//	@Tags        Reference
//	@Accept      json
//	@Produce     json
//	@Param       id   path int                          true "Allergy ID"
//	@Param       body body request.UpdateAllergyRequest true "Updated allergy payload"
//	@Success     200 {object} response.APIResponse[response.EmptyData]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     404 {object} response.APIResponse[response.EmptyData]
//	@Router      /reference/allergies/{id} [put]
func (h *AllergyHandler) UpdateAllergy(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Allergy ID must be a positive integer", nil)
		return
	}

	req, ok := decodeAndValidate[request.UpdateAllergyRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.UpdateAllergyCommand{
		ID:          id,
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
	}

	_, err = messaging.Execute[command.UpdateAllergyCommand, struct{}](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err), "ALLERGY_UPDATE_FAILED", "Could not update allergy", err)
		return
	}

	writeEmptySuccess(w, http.StatusOK, "ALLERGY_UPDATED", "Allergy updated successfully")
}

// DeleteAllergy godoc
//
//	@Summary     Delete a master allergy
//	@Description Removes a reference allergy entry by ID.
//	@Tags        Reference
//	@Produce     json
//	@Param       id path int true "Allergy ID"
//	@Success     200 {object} response.APIResponse[response.EmptyData]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     404 {object} response.APIResponse[response.EmptyData]
//	@Router      /reference/allergies/{id} [delete]
func (h *AllergyHandler) DeleteAllergy(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Allergy ID must be a positive integer", nil)
		return
	}

	cmd := command.DeleteAllergyCommand{ID: id}

	_, err = messaging.Execute[command.DeleteAllergyCommand, struct{}](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err), "ALLERGY_DELETE_FAILED", "Could not delete allergy", err)
		return
	}

	writeEmptySuccess(w, http.StatusOK, "ALLERGY_DELETED", "Allergy deleted successfully")
}

// parseIDParam extracts a named chi URL parameter and parses it as int64.
func parseIDParam(r *http.Request, param string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, param), 10, 64)
}
