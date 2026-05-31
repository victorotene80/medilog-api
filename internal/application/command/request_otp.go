package command

type RequestOTPCommand struct {
	Recipient string
	Purpose   string
	Channel   string
}
