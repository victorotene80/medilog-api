package handler

import (
	"encoding/json"
	"net/http"

	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/request"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
)

type EmergencyContactHandler struct {
	commandBus *messaging.CommandBus
	validator  appContracts.Validator
}

func NewEmergencyContactHandler(commandBus *messaging.CommandBus, validator appContracts.Validator) *EmergencyContactHandler {
	return &EmergencyContactHandler{commandBus: commandBus, validator: validator}
}

// CreateEmergencyContact godoc
//
//	@Summary     Add an emergency contact
//	@Description Creates an emergency contact linked to the authenticated user.
//	@Tags        Emergency Contacts
//	@Accept      json
//	@Produce     json
//	@Security    BearerAuth
//	@Param       body body request.CreateEmergencyContactRequest true "Contact payload"
//	@Success     201 {object} response.APIResponse[struct{}]
//	@Failure     400 {object} response.APIResponse[struct{}]
//	@Failure     401 {object} response.APIResponse[struct{}]
//	@Router      /emergency-contacts/ [post]
func (h *EmergencyContactHandler) CreateEmergencyContact(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	req := new(request.CreateEmergencyContactRequest)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid JSON payload", nil)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "One or more fields are invalid", err.Error())
		return
	}

	cmd := command.EmergencyContactCommand{
		UserID:       userID,
		Name:         req.Name,
		Relationship: req.Relationship,
		Phone:        req.Phone,
		CountryCode:  req.CountryCode,
	}

	result, err := messaging.Execute[command.EmergencyContactCommand, *dto.EmergencyContactDTO](
		h.commandBus, r.Context(), cmd,
	)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "CONTACT_CREATE_FAILED", "Could not create emergency contact", err.Error())
		return
	}

	_ = result // only keep this if you do not need the result

	response.Success(w, http.StatusCreated, "CONTACT_CREATED", "Emergency contact created", &struct{}{})
}
