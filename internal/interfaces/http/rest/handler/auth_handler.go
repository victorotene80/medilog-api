package handler

import (
	"net/http"

	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	domainContracts "github.com/victorotene80/medilog-api/internal/domain/contracts"
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
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     429 {object} response.APIResponse[response.EmptyData]
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
		logAndRespond(w, httperr.StatusFrom(err), "USER_CREATION_FAILED", "Could not create user", err)
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
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Failure     429 {object} response.APIResponse[response.EmptyData]
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
		logAndRespond(w, httperr.StatusFrom(err), "GOOGLE_LOGIN_FAILED", "Google login failed", err)
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
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Failure     429 {object} response.APIResponse[response.EmptyData]
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
		logAndRespond(w, httperr.StatusFrom(err), "LOGIN_FAILED", "Invalid credentials or account locked", err)
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
//	@Success     200 {object} response.APIResponse[response.EmptyData]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     429 {object} response.APIResponse[response.EmptyData]
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
		logAndRespond(w, httperr.StatusFrom(err), "FORGOT_PASSWORD_FAILED", "Could not process password reset request", err)
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
//	@Success     200 {object} response.APIResponse[response.EmptyData]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     429 {object} response.APIResponse[response.EmptyData]
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
		logAndRespond(w, httperr.StatusFrom(err), "RESET_PASSWORD_FAILED", "Could not reset password", err)
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
//	@Success     200 {object} response.APIResponse[response.EmptyData]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Router      /auth/change-password [post]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[request.ChangePasswordRequest](w, r, h.validator)
	if !ok {
		return
	}

	userID, ok := RequireUserID(w, r)
	if !ok {
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
		logAndRespond(w, httperr.StatusFrom(err), "CHANGE_PASSWORD_FAILED", "Could not change password", err)
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
//	@Success     200 {object} response.APIResponse[response.EmptyData]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Router      /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
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
		logAndRespond(w, httperr.StatusFrom(err), "LOGOUT_FAILED", "Could not logout out", err)
		return
	}

	writeEmptySuccess(
		w,
		http.StatusOK,
		"LOGOUT_SUCCESS",
		"Logout successful",
	)
}

func (h *AuthHandler) LoginViaOTP(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[request.LoginViaOTPRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.LoginViaOTPCommand{
		Recipient:         req.Recipient,
		Code:              req.Code,
		Channel:           req.Channel,
		DeviceID:          req.DeviceID,
		DeviceFingerprint: req.DeviceFingerprint,
		DeviceName:        req.DeviceName,
	}

	result, err := messaging.Execute[command.LoginViaOTPCommand, *domainContracts.TokenPair](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err), "LOGIN_OTP_FAILED", "Login via OTP failed", err)
		return
	}

	resp := response.AuthTokensResponse{
		AccessToken:           result.AccessToken.Value,
		AccessTokenExpiresAt:  result.AccessToken.ExpiresAt,
		RefreshToken:          result.RefreshToken.Value,
		RefreshTokenExpiresAt: result.RefreshToken.ExpiresAt,
	}

	response.Success(w, http.StatusOK, "LOGIN_OTP_SUCCESS", "Login successful", &resp)
}

func (h *AuthHandler) RefreshSession(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[request.RefreshSessionRequest](w, r, h.validator)
	if !ok {
		return
	}

	// The struct tag is `omitempty`, so an absent or blank token clears
	// validation. Reject it here rather than paying for a session lookup that
	// can only miss, and so the client sees a malformed-request 400 instead of
	// a 401 that reads as "your session expired".
	if req.RefreshToken == "" {
		response.Error(w, http.StatusBadRequest, "MISSING_REFRESH_TOKEN", "Refresh token is required", nil)
		return
	}

	cmd := command.RefreshSessionCommand{
		RefreshToken: req.RefreshToken,
	}

	result, err := messaging.Execute[command.RefreshSessionCommand, *dto.RefreshResultDTO](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err), "REFRESH_FAILED", "Session refresh failed", err)
		return
	}

	resp := response.AuthTokensResponse{
		AccessToken:           result.AccessToken.Value,
		AccessTokenExpiresAt:  result.AccessToken.ExpiresAt,
		RefreshToken:          result.RefreshToken.Value,
		RefreshTokenExpiresAt: result.RefreshToken.ExpiresAt,
	}

	response.Success(w, http.StatusOK, "REFRESH_SUCCESS", "Session refreshed", &resp)
}

func (h *AuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[request.DeleteAccountRequest](w, r, h.validator)
	if !ok {
		return
	}

	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	cmd := command.DeleteAccountCommand{
		UserID:    userID,
		OTPCode:   req.OTPCode,
		Recipient: req.Recipient,
	}

	_, err := messaging.Execute[command.DeleteAccountCommand, struct{}](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err), "DELETE_ACCOUNT_FAILED", "Could not delete account", err)
		return
	}

	writeEmptySuccess(w, http.StatusOK, "ACCOUNT_DELETED", "Account deleted successfully")
}
