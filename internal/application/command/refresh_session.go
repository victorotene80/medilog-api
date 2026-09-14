package command

// IPAddress and UserAgent come from requestmeta in the handler — see
// LoginViaOTPCommand.
type RefreshSessionCommand struct {
	RefreshToken      string
	DeviceID          string
	DeviceName        string
	DeviceFingerprint string
}
