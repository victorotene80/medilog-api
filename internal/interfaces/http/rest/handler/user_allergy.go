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

type UserAllergyHandler struct {
	commandBus *messaging.CommandBus
	validator  appContracts.Validator
}

func NewUserAllergyHandler(
	commandBus *messaging.CommandBus,
	validator appContracts.Validator,
) *UserAllergyHandler {
	return &UserAllergyHandler{
		commandBus: commandBus,
		validator:  validator,
	}
}

// AddUserAllergies godoc
//
//	@Summary     Add allergies for the authenticated user
//	@Description Records one or more allergies against the authenticated user.
//	             Supports both master-list allergies (pass allergy_id) and
//	             free-text custom allergies (omit allergy_id).
//	@Tags        Health / Allergies
//	@Accept      json
//	@Produce     json
//	@Security    BearerAuth
//	@Param       body body request.AddUserAllergiesRequest true "Allergies payload"
//	@Success     201 {object} response.APIResponse[struct{}]
//	@Failure     400 {object} response.APIResponse[struct{}]
//	@Failure     401 {object} response.APIResponse[struct{}]
//	@Router      /health/allergies/ [post]
func (h *UserAllergyHandler) AddUserAllergies(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	req, ok := decodeAndValidate[request.AddUserAllergiesRequest](w, r, h.validator)
	if !ok {
		return
	}

	items := make([]command.UserAllergyItemCommand, 0, len(req.Allergies))
	for _, item := range req.Allergies {
		items = append(items, command.UserAllergyItemCommand{
			AllergyID:   item.AllergyID,
			Name:        item.Name,
			Description: item.Description,
			Severity:    item.Severity,
			Category:    item.Category,
		})
	}

	cmd := command.CreateUserAllergiesCommand{
		UserID:    userID,
		Allergies: items,
	}

	_, err := messaging.Execute[command.CreateUserAllergiesCommand, struct{}](
		h.commandBus, r.Context(), cmd,
	)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "ALLERGIES_ADD_FAILED", "Could not add allergies", err.Error())
		return
	}

	writeEmptySuccess(w, http.StatusCreated, "ALLERGIES_ADDED", "Allergies recorded successfully")
}

// ListUserAllergies godoc
//
//	@Summary     List authenticated user's allergies
//	@Description Returns all allergies recorded for the authenticated user.
//	             Optionally filter by ?category=1..5
//	@Tags        Health / Allergies
//	@Produce     json
//	@Security    BearerAuth
//	@Param       category query int false "Allergy category (1-5)"
//	@Success     200 {object} response.APIResponse[[]response.UserAllergyResponse]
//	@Failure     401 {object} response.APIResponse[struct{}]
//	@Router      /health/allergies/ [get]
func (h *UserAllergyHandler) ListUserAllergies(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	q := query.ListUserAllergiesQuery{UserID: userID}

	if raw := r.URL.Query().Get("category"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 5 {
			response.Error(w, http.StatusBadRequest, "INVALID_CATEGORY", "category must be an integer between 1 and 5", nil)
			return
		}
		q.Category = &v
	}

	result, err := messaging.Execute[query.ListUserAllergiesQuery, []dto.UserAllergyDTO](
		h.commandBus, r.Context(), q,
	)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "ALLERGIES_FETCH_FAILED", "Could not fetch allergies", err.Error())
		return
	}

	resp := mapper.UserAllergyDTOsToResponse(result)
	response.Success[[]response.UserAllergyResponse](w, http.StatusOK, "ALLERGIES_FETCHED", "Allergies retrieved successfully", &resp)
}

// DeleteUserAllergy godoc
//
//	@Summary     Delete an authenticated user's allergy
//	@Description Removes a specific allergy from the authenticated user's record by public ID.
//	@Tags        Health / Allergies
//	@Produce     json
//	@Security    BearerAuth
//	@Param       publicId path string true "Allergy public ID (UUID)"
//	@Success     200 {object} response.APIResponse[struct{}]
//	@Failure     400 {object} response.APIResponse[struct{}]
//	@Failure     401 {object} response.APIResponse[struct{}]
//	@Failure     404 {object} response.APIResponse[struct{}]
//	@Router      /health/allergies/{publicId} [delete]
func (h *UserAllergyHandler) DeleteUserAllergy(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	publicID := chi.URLParam(r, "publicId")
	if publicID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Allergy public ID is required", nil)
		return
	}

	cmd := command.DeleteUserAllergyCommand{
		UserID:   userID,
		PublicID: publicID,
	}

	_, err := messaging.Execute[command.DeleteUserAllergyCommand, struct{}](
		h.commandBus, r.Context(), cmd,
	)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "ALLERGY_DELETE_FAILED", "Could not delete allergy", err.Error())
		return
	}

	writeEmptySuccess(w, http.StatusOK, "ALLERGY_DELETED", "Allergy removed successfully")
}
