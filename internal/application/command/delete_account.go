package command

type DeleteAccountCommand struct {
	UserID   int64
	OTPCode  string
	Recipient string
}
