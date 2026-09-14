package request

type CreateAIConversationRequest struct {
	Title                     *string `json:"title,omitempty" validate:"omitempty,max=200"`
	RelatedMedicationPublicID *string `json:"related_medication_public_id,omitempty" validate:"omitempty"`
	RelatedVisitPublicID      *string `json:"related_visit_public_id,omitempty" validate:"omitempty"`
}

type UpdateAIConversationRequest struct {
	Title *string `json:"title,omitempty" validate:"omitempty,max=200"`
}

type SendAIMessageRequest struct {
	Message  string `json:"message"  validate:"required,min=1,max=4000"`
	Language string `json:"language" validate:"omitempty,max=50"`
}

