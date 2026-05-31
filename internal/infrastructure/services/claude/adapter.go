package claude

import (
	"context"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
)

var _ appContracts.AIModelService = (*ClaudeModelAdapter)(nil)

type ClaudeModelAdapter struct {
	svc ClaudeService
}

func NewClaudeModelAdapter(svc ClaudeService) *ClaudeModelAdapter {
	return &ClaudeModelAdapter{svc: svc}
}

func (a *ClaudeModelAdapter) Complete(
	ctx context.Context,
	req appContracts.AICompletionRequest,
) (*appContracts.AICompletionResponse, error) {
	msgs := make([]Message, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, Message{
			Role: Role(m.Role),
			Content: []ContentBlock{
				{Type: "text", Text: m.Content},
			},
		})
	}

	claudeReq := MessageRequest{
		System:   req.System,
		Messages: msgs,
	}

	resp, err := a.svc.Send(ctx, claudeReq)
	if err != nil {
		return nil, err
	}

	return &appContracts.AICompletionResponse{
		Text:         resp.Text(),
		PromptTokens: resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
		TotalTokens:  resp.Usage.InputTokens + resp.Usage.OutputTokens,
		Model:        resp.Model,
	}, nil
}