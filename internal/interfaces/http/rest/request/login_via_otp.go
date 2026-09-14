package request

type LoginViaOTPRequest struct {
	Recipient         string `json:"recipient" validate:"required,max=255"`
	Code              string `json:"code" validate:"required"`
	Channel           string `json:"channel" validate:"required,oneof=sms whatsapp"`
	DeviceID          string `json:"device_id"`
	DeviceFingerprint string `json:"device_fingerprint"`
	DeviceName        string `json:"device_name"`
}
