package request

type OTPRequest struct {
	Recipient string `json:"recipient" validate:"required"`
	Channel   string `json:"channel"   validate:"required,oneof=email sms whatsapp"`
	Purpose   string `json:"purpose"   validate:"required"`
}

type VerifyOTPRequest struct {
	Recipient string `json:"recipient" validate:"required"`
	Code      string `json:"code"      validate:"required,min=4,max=8"`
	Channel   string `json:"channel"   validate:"required,oneof=email sms whatsapp"`
	Purpose   string `json:"purpose"   validate:"required"`
}

type VerifyOnboardingOTPRequest struct {
	Recipient string `json:"recipient" validate:"required"`
	Code      string `json:"code" validate:"required"`
	Channel   string `json:"channel" validate:"required"`
	Purpose   string `json:"purpose" validate:"required"`
}
