package aggregates

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/events"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

func intPtr(v int) *int { return &v }

func TestNewAIConversationAggregate(t *testing.T) {
	conv := &entities.AIConversation{
		ID:     1,
		UserID: 10,
		Status: valueobjects.ConversationStatusActive,
	}
	agg := NewAIConversationAggregate(conv)

	assert.Equal(t, int64(1), agg.ID())
	assert.Equal(t, conv, agg.Conversation)
	assert.Empty(t, agg.Messages)
}

func TestAIConversationAddMessage(t *testing.T) {
	conv := &entities.AIConversation{
		ID:     1,
		UserID: 10,
		Status: valueobjects.ConversationStatusActive,
	}
	agg := RestoreAIConversationAggregate(conv, nil, 0)
	now := time.Now().UTC()

	msg := &entities.AIMessage{
		ConversationID: 1,
		ID:             1,
		Role:           valueobjects.AIRoleUser,
		Content:        "What is aspirin?",
	}
	err := agg.AddMessage(msg, now)
	assert.NoError(t, err)
	assert.Len(t, agg.Messages, 1)
	assert.Equal(t, now, *agg.Conversation.LastMessageAt)

	pulled := agg.PullEvents()
	assert.Len(t, pulled, 1)
	assert.Equal(t, events.AIMessageAddedEventName, pulled[0].EventName())
}

func TestAIConversationAddMessageInactive(t *testing.T) {
	conv := &entities.AIConversation{
		ID:     1,
		UserID: 10,
		Status: valueobjects.ConversationStatusArchived,
	}
	agg := RestoreAIConversationAggregate(conv, nil, 0)

	msg := &entities.AIMessage{
		ConversationID: 1,
		ID:             1,
		Role:           valueobjects.AIRoleUser,
		Content:        "Hello",
	}
	err := agg.AddMessage(msg, time.Now().UTC())
	assert.Error(t, err)
	assert.Equal(t, "cannot add message to an inactive conversation", err.Error())
}

func TestAIConversationAddMessageEvent(t *testing.T) {
	conv := &entities.AIConversation{
		ID:     1,
		UserID: 10,
		Status: valueobjects.ConversationStatusActive,
	}
	agg := RestoreAIConversationAggregate(conv, nil, 0)

	msg := &entities.AIMessage{
		ConversationID: 1,
		ID:             1,
		Role:           valueobjects.AIRoleUser,
		Content:        "What is aspirin?",
	}
	err := agg.AddMessage(msg, time.Now().UTC())
	assert.NoError(t, err)

	pulled := agg.PullEvents()
	assert.Len(t, pulled, 1)
	assert.Equal(t, events.AIMessageAddedEventName, pulled[0].EventName())
}

func TestAIConversationBuildContextWindow(t *testing.T) {
	conv := &entities.AIConversation{
		ID:     1,
		UserID: 10,
		Status: valueobjects.ConversationStatusActive,
	}

	msgs := []*entities.AIMessage{
		{ConversationID: 1, ID: 1, Role: valueobjects.AIRoleUser, Content: "msg1", TokenCount: intPtr(10)},
		{ConversationID: 1, ID: 2, Role: valueobjects.AIRoleAssistant, Content: "msg2", TokenCount: intPtr(15)},
		{ConversationID: 1, ID: 3, Role: valueobjects.AIRoleUser, Content: "msg3", TokenCount: intPtr(20)},
	}
	agg := RestoreAIConversationAggregate(conv, msgs, 0)

	window := agg.BuildContextWindow(40)
	assert.Len(t, window, 2)
	assert.Equal(t, "msg2", window[0].Content)
	assert.Equal(t, "msg3", window[1].Content)
}

func TestAIConversationBuildContextWindowWithSummary(t *testing.T) {
	summary := "User discussed aspirin dosage."
	summaryUpTo := int64(2)
	conv := &entities.AIConversation{
		ID:           1,
		UserID:       10,
		Status:       valueobjects.ConversationStatusActive,
		Summary:      &summary,
		SummaryUpTo:  &summaryUpTo,
	}

	msgs := []*entities.AIMessage{
		{ConversationID: 1, ID: 1, Role: valueobjects.AIRoleUser, Content: "old1", TokenCount: intPtr(10)},
		{ConversationID: 1, ID: 2, Role: valueobjects.AIRoleAssistant, Content: "old2", TokenCount: intPtr(10)},
		{ConversationID: 1, ID: 3, Role: valueobjects.AIRoleUser, Content: "new1", TokenCount: intPtr(10)},
	}
	agg := RestoreAIConversationAggregate(conv, msgs, 0)

	window := agg.BuildContextWindow(100)
	assert.Len(t, window, 2)
	assert.Equal(t, valueobjects.AIRoleSystem, window[0].Role)
	assert.Contains(t, window[0].Content, "User discussed aspirin dosage.")
	assert.Equal(t, "new1", window[1].Content)
}

func TestAIConversationTotalTokensUsed(t *testing.T) {
	conv := &entities.AIConversation{
		ID:     1,
		UserID: 10,
		Status: valueobjects.ConversationStatusActive,
	}
	msgs := []*entities.AIMessage{
		{ConversationID: 1, ID: 1, TokenCount: intPtr(10)},
		{ConversationID: 1, ID: 2, TokenCount: intPtr(20)},
		{ConversationID: 1, ID: 3, TokenCount: nil},
	}
	agg := RestoreAIConversationAggregate(conv, msgs, 0)

	assert.Equal(t, 30, agg.TotalTokensUsed())
}

func TestAIConversationShouldSummarise(t *testing.T) {
	conv := &entities.AIConversation{
		ID:     1,
		UserID: 10,
		Status: valueobjects.ConversationStatusActive,
	}
	msgs := []*entities.AIMessage{
		{ConversationID: 1, ID: 1, TokenCount: intPtr(100)},
		{ConversationID: 1, ID: 2, TokenCount: intPtr(50)},
	}
	agg := RestoreAIConversationAggregate(conv, msgs, 0)

	assert.True(t, agg.ShouldSummarise(100))
	assert.False(t, agg.ShouldSummarise(200))
}

func TestAIConversationArchive(t *testing.T) {
	conv := &entities.AIConversation{
		ID:     1,
		UserID: 10,
		Status: valueobjects.ConversationStatusActive,
	}
	agg := RestoreAIConversationAggregate(conv, nil, 0)
	now := time.Now().UTC()

	err := agg.Archive(now)
	assert.NoError(t, err)
	assert.Equal(t, valueobjects.ConversationStatusArchived, agg.Conversation.Status)

	pulled := agg.PullEvents()
	assert.Len(t, pulled, 1)
	assert.Equal(t, events.AIConversationArchivedEventName, pulled[0].EventName())
}

func TestAIConversationArchiveNotActive(t *testing.T) {
	conv := &entities.AIConversation{
		ID:     1,
		UserID: 10,
		Status: valueobjects.ConversationStatusArchived,
	}
	agg := RestoreAIConversationAggregate(conv, nil, 0)

	err := agg.Archive(time.Now().UTC())
	assert.Error(t, err)
	assert.Equal(t, "conversation is not active", err.Error())
}

func TestAIConversationSetTitle(t *testing.T) {
	conv := &entities.AIConversation{
		ID:     1,
		UserID: 10,
		Status: valueobjects.ConversationStatusActive,
	}
	agg := RestoreAIConversationAggregate(conv, nil, 0)
	now := time.Now().UTC()

	agg.SetTitle("Aspirin Dosage Question", now)
	assert.NotNil(t, agg.Conversation.Title)
	assert.Equal(t, "Aspirin Dosage Question", *agg.Conversation.Title)
	assert.Equal(t, now, agg.Conversation.UpdatedAt)
}

func TestAIConversationLastMessageEmpty(t *testing.T) {
	conv := &entities.AIConversation{
		ID:     1,
		UserID: 10,
		Status: valueobjects.ConversationStatusActive,
	}
	agg := RestoreAIConversationAggregate(conv, nil, 0)

	assert.Nil(t, agg.LastMessage())
}

func TestAIConversationLastMessage(t *testing.T) {
	conv := &entities.AIConversation{
		ID:     1,
		UserID: 10,
		Status: valueobjects.ConversationStatusActive,
	}
	msgs := []*entities.AIMessage{
		{ConversationID: 1, ID: 1, Content: "first"},
		{ConversationID: 1, ID: 2, Content: "second"},
		{ConversationID: 1, ID: 3, Content: "third"},
	}
	agg := RestoreAIConversationAggregate(conv, msgs, 0)

	last := agg.LastMessage()
	assert.NotNil(t, last)
	assert.Equal(t, int64(3), last.ID)
	assert.Equal(t, "third", last.Content)
}
