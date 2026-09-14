package request

type LoginRequest struct {
	Email    string `json:"email" validate:"omitempty,email,max=255"`
	Phone    string `json:"phone" validate:"omitempty,max=255"`
	Password string `json:"password" validate:"required,min=8"`
}
