package valueobjects

import "errors"

type OTPChannel string

const (
	OTPChannelEmail    OTPChannel = "email"
	OTPChannelSMS      OTPChannel = "sms"
	OTPChannelWhatsApp OTPChannel = "whatsapp"
)

func NewOTPChannel(raw string) (OTPChannel, error) {
	c := OTPChannel(raw)
	switch c {
	case OTPChannelEmail, OTPChannelSMS, OTPChannelWhatsApp:
		return c, nil
	}
	return "", errors.New("invalid OTP channel: must be email, sms, or whatsapp")
}

func (c OTPChannel) String() string { return string(c) }

type OTPPurpose string

const (
	OTPPurposeEmailVerification OTPPurpose = "email_verification"
	OTPPurposePhoneVerification OTPPurpose = "phone_verification"
	OTPPurposePasswordReset     OTPPurpose = "password_reset"
	OTPPurposeLogin             OTPPurpose = "login"
	OTPPurposeDeleteAccount     OTPPurpose = "delete_account"
)

func NewOTPPurpose(raw string) (OTPPurpose, error) {
	p := OTPPurpose(raw)
	switch p {
	case OTPPurposeEmailVerification,
		OTPPurposePhoneVerification,
		OTPPurposePasswordReset,
		OTPPurposeLogin,
		OTPPurposeDeleteAccount:
		return p, nil
	}
	return "", errors.New("invalid OTP purpose")
}

func (p OTPPurpose) String() string { return string(p) }
