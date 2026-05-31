package command

type ForgotPasswordCommand struct {
	Recipient string // phone or email
}

type ResetPasswordCommand struct {
	Recipient   string
	OTPCode     string
	NewPassword string
}
