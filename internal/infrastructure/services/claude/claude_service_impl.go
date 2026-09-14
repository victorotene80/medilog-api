package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	svchttp "github.com/victorotene80/medilog-api/internal/infrastructure/services/http"
	"github.com/victorotene80/medilog-api/internal/shared/config"
)

const messagesPath = "/v1/messages"

type DefaultClaudeService struct {
	cfg    config.ClaudeConfig
	client svchttp.HTTPService
}

func NewDefaultClaudeService(
	cfg config.ClaudeConfig,
	client svchttp.HTTPService,
) *DefaultClaudeService {
	return &DefaultClaudeService{
		cfg:    cfg,
		client: client,
	}
}

func (s *DefaultClaudeService) Send(
	ctx context.Context,
	req MessageRequest,
) (*MessageResponse, error) {
	// Apply config defaults when the caller leaves fields at zero values.
	if req.Model == "" {
		req.Model = s.cfg.Model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = s.cfg.MaxTokens
	}

	headers := map[string]string{
		"x-api-key":         s.cfg.APIKey,
		"anthropic-version": s.cfg.APIVersion,
	}

	if s.cfg.WorkspaceID != "" {
		headers["anthropic-workspace-id"] = s.cfg.WorkspaceID
	}

	result, err := s.client.Do(ctx, svchttp.HTTPRequest{
		Method:  http.MethodPost,
		URL:     s.cfg.BaseURL + messagesPath,
		Headers: headers,
		Body:    req,
	})
	if err != nil {
		return nil, fmt.Errorf("claude: http error: %w", err)
	}

	if result.StatusCode < 200 || result.StatusCode >= 300 {
		return nil, s.parseAPIError(result)
	}

	var msg MessageResponse
	if err := json.Unmarshal(result.Body, &msg); err != nil {
		return nil, fmt.Errorf("claude: decode response: %w", err)
	}

	return &msg, nil
}

func (s *DefaultClaudeService) Complete(
	ctx context.Context,
	prompt, system string,
) (*MessageResponse, error) {
	req := MessageRequest{
		Messages: []Message{
			{
				Role: RoleUser,
				Content: []ContentBlock{
					{Type: "text", Text: prompt},
				},
			},
		},
	}

	if system != "" {
		req.System = system
	}

	return s.Send(ctx, req)
}

func (s *DefaultClaudeService) parseAPIError(result *svchttp.HTTPResult) error {
	var apiErr APIError
	apiErr.StatusCode = result.StatusCode

	// Best-effort decode; surface raw body if it fails.
	if err := json.Unmarshal(result.Body, &apiErr); err != nil {
		return fmt.Errorf("claude: status %d: %s", result.StatusCode, string(result.Body))
	}

	return fmt.Errorf("claude: status %d (%s): %s",
		apiErr.StatusCode,
		apiErr.Error.Type,
		apiErr.Error.Message,
	)
}
