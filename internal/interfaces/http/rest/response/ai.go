package response

import "time"

// AIConversationResponse identifies a conversation by public UUID only.
//
// related_medication_id and related_visit_id used to be returned as internal
// row ids. They were unusable by clients — every route addresses these by public
// id, and the create request already takes related_medication_public_id — while
// exposing the table's primary-key sequence. Re-adding the association means
// resolving the public ids on the read path, not echoing the internal ones.
type AIConversationResponse struct {
	PublicID      string              `json:"public_id"`
	Title         *string             `json:"title,omitempty"`
	Summary       *string             `json:"summary,omitempty"`
	Status        string              `json:"status"`
	LastMessageAt *time.Time          `json:"last_message_at,omitempty"`
	Messages      []AIMessageResponse `json:"messages,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// AIMessageResponse carries no id: nothing addresses a message individually —
// messages are only ever returned inline within their conversation — so the
// field published the internal row sequence for no use.
type AIMessageResponse struct {
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
	// Quota is the allowance after this message was counted, so the client can
	// update its counter without a follow-up GET /ai/quota.
	Quota AIQuotaResponse `json:"quota"`
}
