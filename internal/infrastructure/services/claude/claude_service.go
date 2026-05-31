package claude

import (
	"context"
)

type ClaudeService interface {
	Send(ctx context.Context, req MessageRequest) (*MessageResponse, error)
	Complete(ctx context.Context, prompt, system string) (*MessageResponse, error)
}
