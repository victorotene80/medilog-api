package handlers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/command"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

type ArchiveAIConversationHandler struct {
	conversations domainRepo.AIConversationRepository
	clock         func() time.Time
}

func NewArchiveAIConversationHandler(
	conversations domainRepo.AIConversationRepository,
	clock func() time.Time,
) *ArchiveAIConversationHandler {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}

	return &ArchiveAIConversationHandler{
		conversations: conversations,
		clock:         clock,
	}
}

func (h *ArchiveAIConversationHandler) Handle(
	ctx context.Context,
	cmd command.ArchiveAIConversationCommand,
) (struct{}, error) {
	agg, err := h.conversations.FindByPublicID(ctx, cmd.UserID, cmd.ConversationPublicID)
	if err != nil {
		return struct{}{}, fmt.Errorf("find ai conversation: %w", err)
	}
	if agg == nil {
		return struct{}{}, errors.New("conversation not found")
	}

	if err := agg.Archive(h.clock()); err != nil {
		return struct{}{}, err
	}

	if err := h.conversations.Update(ctx, agg); err != nil {
		return struct{}{}, fmt.Errorf("archive ai conversation: %w", err)
	}

	return struct{}{}, nil
}
