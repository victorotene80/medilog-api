package request

type ForgotPasswordRequest struct {
	Recipient string `json:"recipient" validate:"required"`
}

type ResetPasswordRequest struct {
	Recipient   string `json:"recipient"    validate:"required"`
	OTPCode     string `json:"otp_code"     validate:"required,len=6"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}
