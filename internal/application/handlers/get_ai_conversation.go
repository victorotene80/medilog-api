package handlers

import (
	"context"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/mapper"
	"github.com/victorotene80/medilog-api/internal/application/query"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

type GetAIConversationHandler struct {
	conversations domainRepo.AIConversationRepository
}

func NewGetAIConversationHandler(
	conversations domainRepo.AIConversationRepository,
) *GetAIConversationHandler {
	return &GetAIConversationHandler{conversations: conversations}
}

func (h *GetAIConversationHandler) Handle(
	ctx context.Context,
	q query.GetAIConversationQuery,
) (*dto.AIConversationDTO, error) {
	agg, err := h.conversations.FindByPublicID(ctx, q.UserID, q.ConversationPublicID)
	if err != nil {
		return nil, fmt.Errorf("find ai conversation: %w", err)
	}
	if agg == nil {
		return nil, nil
	}

	result := mapper.AIConversationAggregateToDTO(agg, true)
	return &result, nil
}
