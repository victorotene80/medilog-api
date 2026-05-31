package contracts

import "context"

type GoogleUserInfo struct {
	GoogleID      string
	Email         string
	EmailVerified bool
	Name          string
	FirstName     string
	LastName      string
	PictureURL    string
}

type GoogleAuthService interface {
	VerifyIDToken(ctx context.Context, idToken string) (*GoogleUserInfo, error)
}