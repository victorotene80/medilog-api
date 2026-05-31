package request

type RefreshSessionRequest struct {
	RefreshToken string `json:"refresh_token,omitempty" validate:"omitempty"`
}
