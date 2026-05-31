package request

type GoogleLoginRequest struct {
	IDToken string `json:"idToken" validate:"required"`
}