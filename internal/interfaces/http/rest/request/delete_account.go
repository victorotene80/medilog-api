package request

type DeleteAccountRequest struct {
	OTPCode   string `json:"otp_code" validate:"required"`
	Recipient string `json:"recipient" validate:"required"`
}
