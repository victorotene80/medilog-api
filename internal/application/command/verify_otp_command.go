package command

type VerifyOTPCommand struct {
	Recipient string
	Code      string
	Channel   string
	Purpose   string
}
