package request

type LogoutRequest struct {
	Reason string `json:"reason,omitempty" validate:"omitempty,max=255"`
}
