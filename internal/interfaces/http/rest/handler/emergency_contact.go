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
//	@Success     201 {object} response.APIResponse[response.EmptyData]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Router      /emergency-contacts/ [post]
func (h *EmergencyContactHandler) CreateEmergencyContact(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	req, ok := decodeAndValidate[request.CreateEmergencyContactRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.EmergencyContactCommand{
		UserID:       userID,
		Name:         req.Name,
		Relationship: req.Relationship,
		Phone:        req.Phone,
		CountryCode:  req.CountryCode,
		IsPrimary:    req.IsPrimary,
	}

	result, err := messaging.Execute[command.EmergencyContactCommand, *dto.EmergencyContactDTO](
		h.commandBus, r.Context(), cmd,
	)
	if err != nil {
		logAndRespond(
			w,
			httperr.StatusFrom(err),
			"CONTACT_CREATE_FAILED",
			"Could not create emergency contact",
			err,
		)
		return
	}

	resp := mapper.EmergencyContactDTOToResponse(result)
	response.Success(w, http.StatusCreated, "CONTACT_CREATED", "Emergency contact created", &resp)
}

// ListEmergencyContacts godoc
//
//	@Summary     List emergency contacts
//	@Description Returns every emergency contact belonging to the authenticated user.
//	@Tags        Emergency Contacts
//	@Produce     json
//	@Security    BearerAuth
//	@Success     200 {object} response.APIResponse[response.ListEmergencyContactsResponse]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Router      /emergency-contacts/ [get]
func (h *EmergencyContactHandler) ListEmergencyContacts(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	result, err := messaging.Execute[query.ListEmergencyContactsQuery, []*dto.EmergencyContactDTO](
		h.commandBus, r.Context(), query.ListEmergencyContactsQuery{UserID: userID},
	)
	if err != nil {
		logAndRespond(
			w,
			httperr.StatusFrom(err),
			"CONTACTS_FETCH_FAILED",
			"Could not list emergency contacts",
			err,
		)
		return
	}

	resp := mapper.EmergencyContactDTOsToResponse(result)
	response.Success(w, http.StatusOK, "CONTACTS_FETCHED", "Emergency contacts retrieved", &resp)
}

// UpdateEmergencyContact godoc
//
//	@Summary     Update an emergency contact
//	@Description Fully replaces the emergency contact identified by its public id.
//	@Tags        Emergency Contacts
//	@Accept      json
//	@Produce     json
//	@Security    BearerAuth
//	@Param       publicId path string true "Emergency contact public id"
//	@Param       body body request.UpdateEmergencyContactRequest true "Contact payload"
//	@Success     200 {object} response.APIResponse[response.EmergencyContactResponse]
//	@Failure     404 {object} response.APIResponse[response.EmptyData]
//	@Failure     409 {object} response.APIResponse[response.EmptyData]
//	@Router      /emergency-contacts/{publicId} [put]
func (h *EmergencyContactHandler) UpdateEmergencyContact(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	publicID, valid := publicIDParam(w, r, "publicId", "Invalid emergency contact id")
	if !valid {
		return
	}

	req, ok := decodeAndValidate[request.UpdateEmergencyContactRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.UpdateEmergencyContactCommand{
		UserID:       userID,
		PublicID:     publicID,
		Name:         req.Name,
		Relationship: req.Relationship,
		Phone:        req.Phone,
		CountryCode:  req.CountryCode,
		IsPrimary:    req.IsPrimary,
	}

	result, err := messaging.Execute[command.UpdateEmergencyContactCommand, *dto.EmergencyContactDTO](
		h.commandBus, r.Context(), cmd,
	)
	if err != nil {
		logAndRespond(
			w,
			httperr.StatusFrom(err),
			"CONTACT_UPDATE_FAILED",
			"Could not update emergency contact",
			err,
		)
		return
	}

	resp := mapper.EmergencyContactDTOToResponse(result)
	response.Success(w, http.StatusOK, "CONTACT_UPDATED", "Emergency contact updated", &resp)
}

// DeleteEmergencyContact godoc
//
//	@Summary     Delete an emergency contact
//	@Description Removes the contact identified by its public id. The user's only
//	             contact cannot be deleted.
//	@Tags        Emergency Contacts
//	@Produce     json
//	@Security    BearerAuth
//	@Param       publicId path string true "Emergency contact public id"
//	@Success     200 {object} response.APIResponse[response.EmptyData]
//	@Failure     404 {object} response.APIResponse[response.EmptyData]
//	@Failure     409 {object} response.APIResponse[response.EmptyData]
//	@Router      /emergency-contacts/{publicId} [delete]
func (h *EmergencyContactHandler) DeleteEmergencyContact(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	publicID, valid := publicIDParam(w, r, "publicId", "Invalid emergency contact id")
	if !valid {
		return
	}

	cmd := command.DeleteEmergencyContactCommand{
		UserID:   userID,
		PublicID: publicID,
	}

	if _, err := messaging.Execute[command.DeleteEmergencyContactCommand, struct{}](
		h.commandBus, r.Context(), cmd,
	); err != nil {
		logAndRespond(
			w,
			httperr.StatusFrom(err),
			"CONTACT_DELETE_FAILED",
			"Could not delete emergency contact",
			err,
		)
		return
	}

	writeEmptySuccess(w, http.StatusOK, "CONTACT_DELETED", "Emergency contact deleted")
}
