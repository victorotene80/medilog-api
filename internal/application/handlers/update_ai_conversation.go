package handlers

import (
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"

	"context"
	"fmt"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/mapper"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

type UpdateAIConversationHandler struct {
	conversations domainRepo.AIConversationRepository
	clock         func() time.Time
}

func NewUpdateAIConversationHandler(
	conversations domainRepo.AIConversationRepository,
	clock func() time.Time,
) *UpdateAIConversationHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}

	return &UpdateAIConversationHandler{
		conversations: conversations,
		clock:         clock,
	}
}

func (h *UpdateAIConversationHandler) Handle(
	ctx context.Context,
	cmd command.UpdateAIConversationCommand,
) (dto.AIConversationDTO, error) {
	if cmd.UserID <= 0 {
		return dto.AIConversationDTO{}, application.NewValidation("user id is required")
	}

	agg, err := h.conversations.FindByPublicID(ctx, cmd.UserID, cmd.ConversationPublicID)
	if err != nil {
		return dto.AIConversationDTO{}, fmt.Errorf("find ai conversation: %w", err)
	}
	if agg == nil {
		return dto.AIConversationDTO{}, application.NewNotFound("conversation not found")
	}

	now := h.clock()
	if cmd.Title != nil {
		trimmed := strings.TrimSpace(*cmd.Title)
		if trimmed == "" {
			agg.Conversation.Title = nil
		} else {
			agg.Conversation.Title = &trimmed
		}
		agg.Conversation.UpdatedAt = now
	}

	if err := h.conversations.Update(ctx, agg); err != nil {
		return dto.AIConversationDTO{}, fmt.Errorf("update ai conversation: %w", err)
	}

	return mapper.AIConversationAggregateToDTO(agg, false), nil
}
