package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type AIMessageRepository interface {
	FindByConversationID(ctx context.Context, conversationID int64) ([]*entities.AIMessage, error)
	FindByConversationIDSince(ctx context.Context, conversationID, afterMessageID int64) ([]*entities.AIMessage, error)
	Save(ctx context.Context, msg *entities.AIMessage) error
	SaveAll(ctx context.Context, msgs []*entities.AIMessage) error
}
