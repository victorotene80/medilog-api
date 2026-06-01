package handler

import (
	"net/http"

	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/httperr"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/mapper"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/request"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
)

type AuthHandler struct {
	commandBus *messaging.CommandBus
	validator  appContracts.Validator
}

func NewAuthHandler(
	commandBus *messaging.CommandBus,
	validator appContracts.Validator,
) *AuthHandler {
	return &AuthHandler{
		commandBus: commandBus,
		validator:  validator,
	}
}

// CreateUser godoc
//
//	@Summary     Register a user
//	@Description Creates a new user account.
//	@Tags        Auth
//	@Accept      json
//	@Produce     json
//	@Param       body body request.RegisterRequest true "Registration payload"
//	@Success     201 {object} response.APIResponse[response.CreateUserResponse]
//	@Failure     400 {object} response.APIResponse[struct{}]
//	@Failure     429 {object} response.APIResponse[struct{}]
//	@Router      /auth/register [post]
func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[request.RegisterRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.RegisterCommand{
		Email:       req.Email,
		Password:    req.Password,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Phone:       req.Phone,
		DOB:         req.DOB,
		Sex:         req.Sex,
		BloodType:   req.BloodType,
		CountryCode: req.CountryCode,
	}

	result, err := messaging.Execute[
		command.RegisterCommand,
		*dto.CreateUserDTO,
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(
			w,
			httperr.StatusFrom(err),
			"USER_CREATION_FAILED",
			"Could not create user",
			err.Error(),
		)
		return
	}

	resp := mapper.CreateUserDTOToResponse(result)

	response.Success(
		w,
		http.StatusCreated,
		"USER_CREATED",
		"User registered successfully",
		&resp,
	)
}

// GoogleLogin godoc
//
//	@Summary     Authenticate with Google
//	@Description Authenticates a user with a Google ID token.
//	@Tags        Auth
//	@Accept      json
//	@Produce     json
//	@Param       body body request.GoogleLoginRequest true "Google login payload"
//	@Success     200 {object} response.APIResponse[response.GoogleAuthResponse]
//	@Failure     400 {object} response.APIResponse[struct{}]
//	@Failure     401 {object} response.APIResponse[struct{}]
//	@Failure     429 {object} response.APIResponse[struct{}]
//	@Router      /auth/google [post]
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[request.GoogleLoginRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.GoogleLoginCommand{
		IDToken: req.IDToken,
	}

	result, err := messaging.Execute[
		command.GoogleLoginCommand,
		*dto.GoogleAuthResultDTO,
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(
			w,
			http.StatusUnauthorized,
			"GOOGLE_LOGIN_FAILED",
			"Google authentication failed",
			err.Error(),
		)
		return
	}

	resp := mapper.GoogleAuthResultDTOToResponse(result)

	response.Success(
		w,
		http.StatusOK,
		"GOOGLE_AUTH_SUCCESS",
		"Google authentication successful",
		&resp,
	)
}

// Login godoc
//
//	@Summary     Log in
//	@Description Logs in with email or phone plus password.
//	@Tags        Auth
//	@Accept      json
//	@Produce     json
//	@Param       body body request.LoginRequest true "Login payload"
//	@Success     200 {object} response.APIResponse[response.LoginResponse]
//	@Failure     400 {object} response.APIResponse[struct{}]
//	@Failure     401 {object} response.APIResponse[struct{}]
//	@Failure     429 {object} response.APIResponse[struct{}]
//	@Router      /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[request.LoginRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := mapper.LoginRequestToCommand(*req)

	result, err := messaging.Execute[
		command.LoginCommand,
		*dto.LoginResultDTO,
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(
			w,
			http.StatusUnauthorized,
			"LOGIN_FAILED",
			"Invalid credentials or account locked",
			err.Error(),
		)
		return
	}

	resp := mapper.LoginResultDTOToResponse(result)

	response.Success(
		w,
		http.StatusOK,
		"LOGIN_SUCCESS",
		"Login successful",
		&resp,
	)
}

// ForgotPassword godoc
//
//	@Summary     Request password reset OTP
//	@Description Sends a password reset OTP when the account can be resolved.
//	@Tags        Auth
//	@Accept      json
//	@Produce     json
//	@Param       body body request.ForgotPasswordRequest true "Forgot password payload"
//	@Success     200 {object} response.APIResponse[struct{}]
//	@Failure     400 {object} response.APIResponse[struct{}]
//	@Failure     429 {object} response.APIResponse[struct{}]
//	@Router      /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[request.ForgotPasswordRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.ForgotPasswordCommand{
		Recipient: req.Recipient,
	}

	_, err := messaging.Execute[
		command.ForgotPasswordCommand,
		struct{},
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"FORGOT_PASSWORD_FAILED",
			"Could not process forgot password request",
			err.Error(),
		)
		return
	}

	writeEmptySuccess(
		w,
		http.StatusOK,
		"PASSWORD_RESET_OTP_SENT",
		"If the account exists, a password reset OTP has been sent",
	)
}

// ResetPassword godoc
//
//	@Summary     Reset password
//	@Description Resets a password with a valid password reset OTP.
//	@Tags        Auth
//	@Accept      json
//	@Produce     json
//	@Param       body body request.ResetPasswordRequest true "Reset password payload"
//	@Success     200 {object} response.APIResponse[struct{}]
//	@Failure     400 {object} response.APIResponse[struct{}]
//	@Failure     429 {object} response.APIResponse[struct{}]
//	@Router      /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[request.ResetPasswordRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.ResetPasswordCommand{
		Recipient:   req.Recipient,
		OTPCode:     req.OTPCode,
		NewPassword: req.NewPassword,
	}

	_, err := messaging.Execute[
		command.ResetPasswordCommand,
		struct{},
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"RESET_PASSWORD_FAILED",
			"Could not reset password",
			err.Error(),
		)
		return
	}

	writeEmptySuccess(
		w,
		http.StatusOK,
		"PASSWORD_RESET_SUCCESS",
		"Password reset successfully",
	)
}

// ChangePassword godoc
//
//	@Summary     Change password
//	@Description Changes the authenticated user's password.
//	@Tags        Auth
//	@Accept      json
//	@Produce     json
//	@Security    BearerAuth
//	@Param       body body request.ChangePasswordRequest true "Change password payload"
//	@Success     200 {object} response.APIResponse[struct{}]
//	@Failure     400 {object} response.APIResponse[struct{}]
//	@Failure     401 {object} response.APIResponse[struct{}]
//	@Router      /auth/change-password [post]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[request.ChangePasswordRequest](w, r, h.validator)
	if !ok {
		return
	}

	userID, ok := UserIDFrom(r.Context())
	if !ok || userID <= 0 {
		response.Error(
			w,
			http.StatusUnauthorized,
			"UNAUTHENTICATED",
			"Authentication is required",
			nil,
		)
		return
	}

	cmd := command.ChangePasswordCommand{
		UserID:      userID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}

	_, err := messaging.Execute[
		command.ChangePasswordCommand,
		struct{},
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"CHANGE_PASSWORD_FAILED",
			"Could not change password",
			err.Error(),
		)
		return
	}

	writeEmptySuccess(
		w,
		http.StatusOK,
		"PASSWORD_CHANGED",
		"Password changed successfully",
	)
}

// Logout godoc
//
//	@Summary     Log out
//	@Description Revokes the current authenticated session.
//	@Tags        Auth
//	@Produce     json
//	@Security    BearerAuth
//	@Success     200 {object} response.APIResponse[struct{}]
//	@Failure     400 {object} response.APIResponse[struct{}]
//	@Failure     401 {object} response.APIResponse[struct{}]
//	@Router      /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok || userID <= 0 {
		response.Error(
			w,
			http.StatusUnauthorized,
			"UNAUTHENTICATED",
			"Authentication is required",
			nil,
		)
		return
	}

	sessionID, ok := SessionIDFrom(r.Context())
	if !ok || sessionID == "" {
		response.Error(
			w,
			http.StatusUnauthorized,
			"UNAUTHENTICATED",
			"Session context is required",
			nil,
		)
		return
	}

	cmd := command.LogoutCommand{
		UserID:    userID,
		SessionID: sessionID,
	}

	_, err := messaging.Execute[
		command.LogoutCommand,
		struct{},
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"LOGOUT_FAILED",
			"Could not logout user",
			err.Error(),
		)
		return
	}

	writeEmptySuccess(
		w,
		http.StatusOK,
		"LOGOUT_SUCCESS",
		"Logout successful",
	)
}
