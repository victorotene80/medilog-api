package dto

import "time"

type AIConversationDTO struct {
	ID                  int64
	PublicID            string
	UserID              int64
	Title               *string
	RelatedMedicationID *int64
	RelatedVisitID      *int64
	Summary             *string
	Status              string
	LastMessageAt       *time.Time
	Messages            []AIMessageDTO
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type AIMessageDTO struct {
	ID             int64
	ConversationID int64
	Role           string
	Content        string
	TokenCount     *int
	Meta           map[string]any
	CreatedAt      time.Time
}

type SendAIMessageResultDTO struct {
	Conversation AIConversationDTO
	Message      AIMessageDTO
	Reply        AIMessageDTO
	Model        string
	PromptTokens int
	OutputTokens int
	ContextMeta  map[string]any
	// Quota reflects the allowance *after* this message was counted, so the
	// client can update its counter without a follow-up GET /ai/quota.
	Quota AIQuotaDTO
}

type PatientContextChatRequest struct {
	History             []ChatMessage
	PatientContext      string
	ConversationSummary *string
	Language            string
}

type PatientContextChatResponse struct {
	Reply        string
	Model        string
	PromptTokens int
	OutputTokens int
	TotalTokens  int
}
