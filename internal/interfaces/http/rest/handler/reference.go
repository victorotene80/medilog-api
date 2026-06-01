package handler

import (
	"net/http"
	"strconv"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/mapper"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
)

// ListCountriesQuery is an empty marker used to dispatch the no-arg handler
// through the command bus.

type ReferenceHandler struct {
	commandBus *messaging.CommandBus
	validator  appContracts.Validator
}

func NewReferenceHandler(commandBus *messaging.CommandBus, validator appContracts.Validator) *ReferenceHandler {
	return &ReferenceHandler{commandBus: commandBus, validator: validator}
}

// ListCountries godoc
//
//	@Summary     List all supported countries
//	@Description Returns a list of countries with ISO codes, dial codes and flag URLs.
//	             Public — no authentication required.
//	@Tags        Reference Data
//	@Produce     json
//	@Success     200 {object} response.APIResponse[[]response.CountryResponse]
//	@Failure     500 {object} response.APIResponse[struct{}]
//	@Router      /reference/countries [get]
func (h *ReferenceHandler) ListCountries(w http.ResponseWriter, r *http.Request) {
	results, err := messaging.Execute[query.GetCountriesQuery, []dto.CountryDTO](
		h.commandBus, r.Context(), query.GetCountriesQuery{},
	)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "COUNTRIES_FETCH_FAILED", "Could not retrieve countries", err.Error())
		return
	}

	countries := mapper.CountryDTOsToResponse(results)
	response.Success(w, http.StatusOK, "COUNTRIES_FETCHED", "Countries retrieved", &countries)
}

// ListAllergies is kept for compatibility with older wiring. The active route
// is served by AllergyHandler.ListAllergies in router.go.
func (h *ReferenceHandler) ListAllergies(w http.ResponseWriter, r *http.Request) {
	var category *int
	if rawCategory := r.URL.Query().Get("category"); rawCategory != "" {
		parsed, err := strconv.Atoi(rawCategory)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "INVALID_CATEGORY", "Category must be a number", nil)
			return
		}
		category = &parsed
	}

	results, err := messaging.Execute[query.GetAllergiesQuery, []dto.AllergyDTO](
		h.commandBus,
		r.Context(),
		query.GetAllergiesQuery{Category: category},
	)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "ALLERGIES_FETCH_FAILED", "Could not retrieve allergies", err.Error())
		return
	}

	allergies := mapper.AllergyDTOsToResponse(results)
	response.Success(w, http.StatusOK, "ALLERGIES_FETCHED", "Allergies retrieved", &allergies)
}
