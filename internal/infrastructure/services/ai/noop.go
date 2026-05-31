package ai

import (
	"context"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application/contracts"
)

var _ contracts.AIModelService = (*NoopModelService)(nil)

type NoopModelService struct {
	reason string
}

func NewNoopModelService(reason string) *NoopModelService {
	return &NoopModelService{reason: reason}
}

func (s *NoopModelService) Complete(
	ctx context.Context,
	req contracts.AICompletionRequest,
) (*contracts.AICompletionResponse, error) {
	_ = ctx
	_ = req

	if s.reason == "" {
		return nil, fmt.Errorf("ai model is not configured")
	}

	return nil, fmt.Errorf("ai model is not configured: %s", s.reason)
}
