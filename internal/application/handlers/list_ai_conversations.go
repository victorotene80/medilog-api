package handlers

import (
	"context"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/mapper"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

type ListAIConversationsHandler struct {
	conversations domainRepo.AIConversationRepository
}

func NewListAIConversationsHandler(
	conversations domainRepo.AIConversationRepository,
) *ListAIConversationsHandler {
	return &ListAIConversationsHandler{conversations: conversations}
}

func (h *ListAIConversationsHandler) Handle(
	ctx context.Context,
	q query.ListAIConversationsQuery,
) ([]dto.AIConversationDTO, error) {
	aggs, err := h.findConversations(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list ai conversations: %w", err)
	}

	return mapper.AIConversationAggregatesToDTO(aggs, false), nil
}

func (h *ListAIConversationsHandler) findConversations(
	ctx context.Context,
	q query.ListAIConversationsQuery,
) ([]*aggregates.AIConversationAggregate, error) {
	if q.ActiveOnly {
		return h.conversations.FindActiveByUserID(ctx, q.UserID)
	}
	return h.conversations.FindByUserID(ctx, q.UserID)
}
