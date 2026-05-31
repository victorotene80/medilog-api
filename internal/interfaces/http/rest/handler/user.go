package handler

import (
	"net/http"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/mapper"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
)

type UserHandler struct {
	commandBus *messaging.CommandBus
	validator  appContracts.Validator
}

func NewUserHandler(commandBus *messaging.CommandBus, validator appContracts.Validator) *UserHandler {
	return &UserHandler{commandBus: commandBus, validator: validator}
}

// GetMe godoc
//
//	@Summary     Get authenticated user's profile
//	@Description Returns the profile of the currently authenticated user.
//	             The user ID is extracted from the Bearer token — no path param needed.
//	@Tags        Users
//	@Produce     json
//	@Security    BearerAuth
//	@Success     200 {object} response.APIResponse[response.GetUserResponse]
//	@Failure     401 {object} response.APIResponse[struct{}]
//	@Failure     404 {object} response.APIResponse[struct{}]
//	@Router      /users/me [get]
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	q := query.GetUserQuery{ID: userID}

	result, err := messaging.Execute[query.GetUserQuery, *dto.GetUserDTO](
		h.commandBus, r.Context(), q,
	)
	if err != nil {
		response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found", err.Error())
		return
	}
	if result == nil {
		response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found", nil)
		return
	}

	resp := mapper.GetUserDTOToResponse(result)
	response.Success(w, http.StatusOK, "USER_FETCHED", "User profile retrieved", &resp)
}
