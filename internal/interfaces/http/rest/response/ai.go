package response

import "time"

type AIConversationResponse struct {
	PublicID            string              `json:"public_id"`
	Title               *string             `json:"title,omitempty"`
	RelatedMedicationID *int64              `json:"related_medication_id,omitempty"`
	RelatedVisitID      *int64              `json:"related_visit_id,omitempty"`
	Summary             *string             `json:"summary,omitempty"`
	Status              string              `json:"status"`
	LastMessageAt       *time.Time          `json:"last_message_at,omitempty"`
	Messages            []AIMessageResponse `json:"messages,omitempty"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
}

type AIMessageResponse struct {
	ID        int64          `json:"id"`
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	Meta      map[string]any `json:"meta,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type SendAIMessageResponse struct {
	Conversation AIConversationResponse `json:"conversation"`
	Message      AIMessageResponse      `json:"message"`
	Reply        AIMessageResponse      `json:"reply"`
	Model        string                 `json:"model"`
	PromptTokens int                    `json:"prompt_tokens"`
	OutputTokens int                    `json:"output_tokens"`
	ContextMeta  map[string]any         `json:"context_meta,omitempty"`
}
