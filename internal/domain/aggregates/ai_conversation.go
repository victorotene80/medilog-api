package aggregates

import (
	"errors"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/events/types"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type AIConversationAggregate struct {
	*AggregateRoot
	Conversation *entities.AIConversation
	Messages     []*entities.AIMessage
}

func NewAIConversationAggregate(c *entities.AIConversation) *AIConversationAggregate {
	return &AIConversationAggregate{
		AggregateRoot: NewAggregateRoot(c.ID, 0),
		Conversation:  c,
		Messages:      make([]*entities.AIMessage, 0),
	}
}

func RestoreAIConversationAggregate(
	c *entities.AIConversation,
	messages []*entities.AIMessage,
	version int,
) *AIConversationAggregate {
	return &AIConversationAggregate{
		AggregateRoot: NewAggregateRoot(c.ID, version),
		Conversation:  c,
		Messages:      messages,
	}
}

func (a *AIConversationAggregate) AddMessage(msg *entities.AIMessage, now time.Time) error {
	if !a.Conversation.IsActive() {
		return errors.New("cannot add message to an inactive conversation")
	}
	a.Messages = append(a.Messages, msg)
	a.Conversation.TouchLastMessage(now)
	a.RaiseEvent(types.NewAIMessageAddedEvent(a.Conversation.ID, string(msg.Role)))
	return nil
}

// BuildContextWindow returns the messages to send to the AI API.
// It respects the token budget and prepends the rolling summary when present.
// Messages are returned oldest-first as required by the API.
func (a *AIConversationAggregate) BuildContextWindow(maxTokens int) []*entities.AIMessage {
	var window []*entities.AIMessage
	budget := maxTokens

	// Walk from newest to oldest, accumulating until budget is exhausted
	for i := len(a.Messages) - 1; i >= 0; i-- {
		msg := a.Messages[i]

		// Skip messages already covered by summary
		if a.Conversation.SummaryUpTo != nil && msg.ID <= *a.Conversation.SummaryUpTo {
			break
		}

		tokens := 0
		if msg.TokenCount != nil {
			tokens = *msg.TokenCount
		}

		if budget-tokens < 0 {
			break
		}

		budget -= tokens
		window = append([]*entities.AIMessage{msg}, window...)
	}

	// Prepend summary as a system message if one exists
	if a.Conversation.Summary != nil {
		summary := &entities.AIMessage{
			ConversationID: a.Conversation.ID,
			Role:           valueobjects.AIRoleSystem,
			Content:        "Conversation summary so far:\n" + *a.Conversation.Summary,
			CreatedAt:      time.Time{},
		}
		window = append([]*entities.AIMessage{summary}, window...)
	}

	return window
}

func (a *AIConversationAggregate) TotalTokensUsed() int {
	total := 0
	for _, m := range a.Messages {
		if m.TokenCount != nil {
			total += *m.TokenCount
		}
	}
	return total
}

// ShouldSummarise returns true when accumulated tokens exceed the threshold.
func (a *AIConversationAggregate) ShouldSummarise(tokenThreshold int) bool {
	return a.TotalTokensUsed() >= tokenThreshold
}

// ApplySummary stores a new rolling summary up to the given message ID.
func (a *AIConversationAggregate) ApplySummary(summary string, upToMessageID int64, now time.Time) {
	a.Conversation.UpdateSummary(summary, upToMessageID, now)
}

// Archive closes the conversation.
func (a *AIConversationAggregate) Archive(now time.Time) error {
	if !a.Conversation.IsActive() {
		return errors.New("conversation is not active")
	}
	a.Conversation.Archive(now)
	a.RaiseEvent(types.NewAIConversationArchivedEvent(a.Conversation.ID))
	return nil
}

// SetTitle auto-sets the conversation title (typically from first message).
func (a *AIConversationAggregate) SetTitle(title string, now time.Time) {
	a.Conversation.Title = &title
	a.Conversation.UpdatedAt = now
}

// LastMessage returns the most recent message or nil.
func (a *AIConversationAggregate) LastMessage() *entities.AIMessage {
	if len(a.Messages) == 0 {
		return nil
	}
	return a.Messages[len(a.Messages)-1]
}
