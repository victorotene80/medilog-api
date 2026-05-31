package types

import "github.com/victorotene80/medilog-api/internal/domain/events"

type AIConversationCreatedPayload struct {
	UserID int64 `json:"user_id"`
}

type AIMessageAddedPayload struct {
	Sender string `json:"sender"`
}

type AIConversationArchivedPayload struct{}

func NewAIConversationCreatedEvent(conversationID int64, userID int64) events.DomainEvent {
	return events.NewEvent(
		events.AIConversationCreatedEventName,
		conversationID,
		AIConversationCreatedPayload{UserID: userID},
		nil,
	)
}

func NewAIMessageAddedEvent(conversationID int64, sender string) events.DomainEvent {
	return events.NewEvent(
		events.AIMessageAddedEventName,
		conversationID,
		AIMessageAddedPayload{Sender: sender},
		nil,
	)
}

func NewAIConversationArchivedEvent(conversationID int64) events.DomainEvent {
	return events.NewEvent(events.AIConversationArchivedEventName, conversationID, AIConversationArchivedPayload{}, nil)
}