package command

// IPAddress and UserAgent are deliberately absent: the handler reads them from
// requestmeta, which carries the value normalized by chi's RealIP. Passing
// r.RemoteAddr here wrote "203.0.113.9:54321" into users.last_login_ip and
// refresh_tokens.ip_address while password login wrote "203.0.113.9", leaving
// the audit trail with two formats for one concept.
type LoginViaOTPCommand struct {
	Recipient         string
	Code              string
	Channel           string
	DeviceID          string
	DeviceFingerprint string
	DeviceName        string
}
