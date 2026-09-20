package contracts

import "context"

type AIMessage struct {
	Role    string
	Content string
}

type AICompletionRequest struct {
	System   string
	Messages []AIMessage
	// MaxTokens overrides the configured default completion budget when set.
	// Zero means "use the configured default".
	MaxTokens int
}

type AICompletionResponse struct {
	Text         string
	PromptTokens int
	OutputTokens int
	TotalTokens  int
	Model        string
}

type AIModelService interface {
	Complete(ctx context.Context, req AICompletionRequest) (*AICompletionResponse, error)
} 