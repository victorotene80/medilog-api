package command

type VerifyOnboardingOTPCommand struct {
	Recipient string
	Channel   string
	Purpose   string
	Code      string
}
