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
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Failure     404 {object} response.APIResponse[response.EmptyData]
//	@Router      /users/me [get]
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	q := query.GetUserQuery{ID: userID}

	result, err := messaging.Execute[query.GetUserQuery, *dto.GetUserDTO](
		h.commandBus, r.Context(), q,
	)
	if err != nil {
		logAndRespond(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found", err)
		return
	}
	if result == nil {
		response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found", nil)
		return
	}

	resp := mapper.GetUserDTOToResponse(result)
	response.Success(w, http.StatusOK, "USER_FETCHED", "User profile retrieved", &resp)
}

// UpdateMe godoc
//
//	@Summary     Update the authenticated user's profile
//	@Description Partial update — omitted fields are left unchanged. Returns the
//	             same shape as GET /users/me.
//
//	             Email and phone cannot be changed here: they require OTP
//	             re-verification, and because unknown JSON fields are rejected, a
//	             body containing either returns 400 rather than ignoring it.
//	@Tags        Users
//	@Accept      json
//	@Produce     json
//	@Security    BearerAuth
//	@Param       body body request.UpdateUserRequest true "Fields to update"
//	@Success     200 {object} response.APIResponse[response.GetUserResponse]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Failure     404 {object} response.APIResponse[response.EmptyData]
//	@Failure     422 {object} response.APIResponse[response.EmptyData]
//	@Router      /users/me [patch]
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	req, ok := decodeAndValidate[request.UpdateUserRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.UpdateUserCommand{
		UserID:          userID,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		DOB:             req.DOB,
		Sex:             req.Sex,
		BloodType:       req.BloodType,
		AvatarURL:       req.AvatarURL,
		CountryCode:     req.CountryCode,
		Height:          req.Height,
		Weight:          req.Weight,
		WeightUnit:      req.WeightUnit,
		TemperatureUnit: req.TemperatureUnit,
	}

	result, err := messaging.Execute[command.UpdateUserCommand, *dto.GetUserDTO](
		h.commandBus, r.Context(), cmd,
	)
	if err != nil {
		logAndRespond(
			w,
			httperr.StatusFrom(err),
			"USER_UPDATE_FAILED",
			"Could not update profile",
			err,
		)
		return
	}

	if result == nil {
		response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found", nil)
		return
	}

	resp := mapper.GetUserDTOToResponse(result)
	response.Success(w, http.StatusOK, "USER_UPDATED", "Profile updated", &resp)
}

// GetNotificationPreferences godoc
//
//	@Summary     Get notification preferences
//	@Description Returns the caller's delivery preferences and reminder timezone.
//	@Tags        Users
//	@Produce     json
//	@Security    BearerAuth
//	@Success     200 {object} response.APIResponse[response.NotificationPreferencesResponse]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Failure     404 {object} response.APIResponse[response.EmptyData]
//	@Router      /users/me/notification-preferences [get]
func (h *UserHandler) GetNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	result, err := messaging.Execute[
		query.GetNotificationPreferencesQuery,
		*dto.NotificationPreferencesDTO,
	](h.commandBus, r.Context(), query.GetNotificationPreferencesQuery{UserID: userID})
	if err != nil {
		logAndRespond(
			w,
			httperr.StatusFrom(err),
			"PREFERENCES_FETCH_FAILED",
			"Could not retrieve notification preferences",
			err,
		)
		return
	}

	if result == nil {
		response.Error(w, http.StatusNotFound, "USER_PROFILE_NOT_FOUND", "User profile not found", nil)
		return
	}

	resp := mapper.NotificationPreferencesDTOToResponse(*result)
	response.Success(w, http.StatusOK, "PREFERENCES_FETCHED", "Notification preferences retrieved", &resp)
}

// UpdateNotificationPreferences godoc
//
//	@Summary     Update notification preferences
//	@Description Partial update — omitted toggles are left unchanged. The
//	             timezone is an IANA name and is validated on write, because the
//	             reminder scheduler resolves it on every tick.
//	@Tags        Users
//	@Accept      json
//	@Produce     json
//	@Security    BearerAuth
//	@Param       body body request.UpdateNotificationPreferencesRequest true "Preferences to update"
//	@Success     200 {object} response.APIResponse[response.NotificationPreferencesResponse]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Failure     404 {object} response.APIResponse[response.EmptyData]
//	@Failure     422 {object} response.APIResponse[response.EmptyData]
//	@Router      /users/me/notification-preferences [patch]
func (h *UserHandler) UpdateNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	req, ok := decodeAndValidate[request.UpdateNotificationPreferencesRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.UpdateNotificationPreferencesCommand{
		UserID:                      userID,
		MedicationRemindersEnabled:  req.MedicationRemindersEnabled,
		RefillRemindersEnabled:      req.RefillRemindersEnabled,
		AppointmentRemindersEnabled: req.AppointmentRemindersEnabled,
		AIHealthTipsEnabled:         req.AIHealthTipsEnabled,
		SupportUpdatesEnabled:       req.SupportUpdatesEnabled,
		AppUpdatesEnabled:           req.AppUpdatesEnabled,
		PushEnabled:                 req.PushEnabled,
		EmailEnabled:                req.EmailEnabled,
		SMSEnabled:                  req.SMSEnabled,
		WhatsAppEnabled:             req.WhatsAppEnabled,
		Timezone:                    req.Timezone,
	}

	result, err := messaging.Execute[
		command.UpdateNotificationPreferencesCommand,
		*dto.NotificationPreferencesDTO,
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(
			w,
			httperr.StatusFrom(err),
			"PREFERENCES_UPDATE_FAILED",
			"Could not update notification preferences",
			err,
		)
		return
	}

	if result == nil {
		response.Error(w, http.StatusNotFound, "USER_PROFILE_NOT_FOUND", "User profile not found", nil)
		return
	}

	resp := mapper.NotificationPreferencesDTOToResponse(*result)
	response.Success(w, http.StatusOK, "PREFERENCES_UPDATED", "Notification preferences updated", &resp)
}
