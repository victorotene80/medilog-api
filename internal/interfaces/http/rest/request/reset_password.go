package request

type ForgotPasswordRequest struct {
	Recipient string `json:"recipient" validate:"required,max=255"`
}

type ResetPasswordRequest struct {
	Recipient   string `json:"recipient"    validate:"required,max=255"`
	OTPCode     string `json:"otp_code"     validate:"required,len=6"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}
