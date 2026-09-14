package handler

import (
	"encoding/json"
	"net/http"

	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/mapper"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/request"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
)

type OTPHandler struct {
	commandBus *messaging.CommandBus
	validator  appContracts.Validator
}

func NewOTPHandler(commandBus *messaging.CommandBus, validator appContracts.Validator) *OTPHandler {
	return &OTPHandler{commandBus: commandBus, validator: validator}
}

// RequestOTP godoc
//
//	@Summary     Request an OTP
//	@Description Sends a one-time password to the given recipient via the chosen channel.
//	             No auth required — used for email/phone verification and password reset.
//	@Tags        Auth / OTP
//	@Accept      json
//	@Produce     json
//	@Param       body body request.OTPRequest true "OTP request payload"
//	@Success     200 {object} response.APIResponse[response.RequestOTPResponse]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     422 {object} response.APIResponse[response.EmptyData]
//	@Router      /auth/otp/request [post]
func (h *OTPHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	req := new(request.OTPRequest)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid JSON payload", nil)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		logAndRespond(w, http.StatusBadRequest, "VALIDATION_ERROR", "One or more fields are invalid", err)
		return
	}

	cmd := command.RequestOTPCommand{
		Recipient: req.Recipient,
		Channel:   req.Channel,
		Purpose:   req.Purpose,
	}

	result, err := messaging.Execute[command.RequestOTPCommand, *dto.RequestOTPResultDTO](
		h.commandBus, r.Context(), cmd,
	)
	if err != nil {
		logAndRespond(w, http.StatusUnprocessableEntity, "OTP_REQUEST_FAILED", "Could not send OTP", err)
		return
	}

	resp := mapper.RequestOTPResultDTOToResponse(result)
	response.Success(w, http.StatusOK, "OTP_SENT", "OTP sent successfully", &resp)
}

// VerifyOTP godoc
//
//	@Summary     Verify an OTP
//	@Description Validates the submitted OTP code and marks the channel (email/phone) as verified.
//	             No auth required.
//	@Tags        Auth / OTP
//	@Accept      json
//	@Produce     json
//	@Param       body body request.VerifyOTPRequest true "OTP verify payload"
//	@Success     200 {object} response.APIResponse[response.VerifyOTPResponse]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     422 {object} response.APIResponse[response.EmptyData]
//	@Router      /auth/otp/verify [post]
func (h *OTPHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	req := new(request.VerifyOTPRequest)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid JSON payload", nil)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		logAndRespond(w, http.StatusBadRequest, "VALIDATION_ERROR", "One or more fields are invalid", err)
		return
	}

	cmd := command.VerifyOTPCommand{
		Recipient: req.Recipient,
		Code:      req.Code,
		Channel:   req.Channel,
		Purpose:   req.Purpose,
	}

	result, err := messaging.Execute[command.VerifyOTPCommand, *dto.VerifyOTPResultDTO](
		h.commandBus, r.Context(), cmd,
	)
	if err != nil {
		logAndRespond(w, http.StatusUnprocessableEntity, "OTP_VERIFY_FAILED", "Invalid or expired OTP", err)
		return
	}

	resp := mapper.VerifyOTPResultDTOToResponse(result)
	response.Success(w, http.StatusOK, "OTP_VERIFIED", "OTP verified successfully", &resp)
}

// VerifyOnboardingOTP godoc
//
//	@Summary     Verify onboarding OTP
//	@Description Verifies the OTP submitted during registration/onboarding.
//	             If valid, it marks the user email/phone as verified, activates the account,
//	             creates a session, and returns access/refresh tokens.
//	             No auth required.
//	@Tags        Auth / OTP
//	@Accept      json
//	@Produce     json
//	@Param       body body request.VerifyOnboardingOTPRequest true "Onboarding OTP verify payload"
//	@Success     200 {object} response.APIResponse[response.VerifyOnboardingOTPResponse]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     422 {object} response.APIResponse[response.EmptyData]
//	@Router      /auth/otp/verify-onboarding [post]
func (h *OTPHandler) VerifyOnboardingOTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	req := new(request.VerifyOnboardingOTPRequest)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(req); err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST_BODY",
			"Invalid JSON payload",
			nil,
		)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		logAndRespond(
			w,
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"One or more fields are invalid",
			err,
		)
		return
	}

	cmd := command.VerifyOnboardingOTPCommand{
		Recipient: req.Recipient,
		Code:      req.Code,
		Channel:   req.Channel,
		Purpose:   req.Purpose,
	}

	result, err := messaging.Execute[
		command.VerifyOnboardingOTPCommand,
		*dto.VerifyOnboardingOTPResultDTO,
	](
		h.commandBus,
		r.Context(),
		cmd,
	)
	if err != nil {
		logAndRespond(
			w,
			http.StatusUnprocessableEntity,
			"ONBOARDING_OTP_VERIFY_FAILED",
			"Invalid or expired OTP",
			err,
		)
		return
	}

	resp := mapper.VerifyOnboardingOTPResultDTOToResponse(result)

	response.Success(
		w,
		http.StatusOK,
		"ONBOARDING_OTP_VERIFIED",
		"OTP verified successfully",
		&resp,
	)
}
