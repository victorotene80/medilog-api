package request

type RegisterRequest struct {
	FirstName   string  `json:"first_name" validate:"required,min=2,max=50"`
	LastName    string  `json:"last_name" validate:"required,min=2,max=50"`
	Email       *string `json:"email" validate:"omitempty,email"`
	Phone       *string `json:"phone" validate:"omitempty,min=8,max=15"`
	Password    string  `json:"password" validate:"required,min=8"`
	DOB         string  `json:"dob" validate:"required"`
	Sex         int     `json:"sex" validate:"required"`
	BloodType   string  `json:"blood_type" validate:"required,max=3"`
	CountryCode string  `json:"country_code" validate:"required,min=1,max=4"`
}
