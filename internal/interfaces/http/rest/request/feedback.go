package request

type SubmitFeedbackRequest struct {
	Rating      *int    `json:"rating" validate:"omitempty,min=1,max=5"`
	Title       *string `json:"title" validate:"omitempty,max=500"`
	Message     string  `json:"message" validate:"required"`
	AppVersion  *string `json:"app_version" validate:"omitempty,max=50"`
	Platform    *string `json:"platform" validate:"omitempty,max=50"`
	DeviceModel *string `json:"device_model" validate:"omitempty,max=200"`
}
