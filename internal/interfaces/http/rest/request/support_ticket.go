package request

type CreateSupportTicketRequest struct {
	Subject  string `json:"subject" validate:"required"`
	Category string `json:"category" validate:"required,oneof=general billing technical feature_request bug_report"`
	Priority string `json:"priority" validate:"required,oneof=low normal high urgent"`
	Message  string `json:"message" validate:"required"`
}

type AddSupportMessageRequest struct {
	Message string `json:"message" validate:"required"`
}
