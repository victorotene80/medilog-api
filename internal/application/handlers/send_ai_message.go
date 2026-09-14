package handlers

import (
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"

	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/mapper"
	appServices "github.com/victorotene80/medilog-api/internal/application/services"
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type SendAIMessageHandler struct {
	conversations         domainRepo.AIConversationRepository
	messages              domainRepo.AIMessageRepository
	profiles              domainRepo.UserProfileRepository
	contextBuilder        *appServices.AIContextBuilder
	ai                    *appServices.AIService
	clock                 func() time.Time
	maxContextTokens      int
	maxContextMessages    int
	summaryTokenThreshold int
}

func NewSendAIMessageHandler(
	conversations domainRepo.AIConversationRepository,
	messages domainRepo.AIMessageRepository,
	profiles domainRepo.UserProfileRepository,
	contextBuilder *appServices.AIContextBuilder,
	ai *appServices.AIService,
	clock func() time.Time,
	maxContextTokens int,
	maxContextMessages int,
	summaryTokenThreshold int,
) *SendAIMessageHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}

	return &SendAIMessageHandler{
		conversations:         conversations,
		messages:              messages,
		profiles:              profiles,
		contextBuilder:        contextBuilder,
		ai:                    ai,
		clock:                 clock,
		maxContextTokens:      maxContextTokens,
		maxContextMessages:    maxContextMessages,
		summaryTokenThreshold: summaryTokenThreshold,
	}
}

func (h *SendAIMessageHandler) Handle(
	ctx context.Context,
	cmd command.SendAIMessageCommand,
) (*dto.SendAIMessageResultDTO, error) {
	content := strings.TrimSpace(cmd.Message)
	if content == "" {
		return nil, application.NewValidation("message is required")
	}
	if h.contextBuilder == nil {
		return nil, errors.New("ai context builder is not configured")
	}
	if h.ai == nil {
		return nil, errors.New("ai service is not configured")
	}

	agg, err := h.conversations.FindByPublicID(ctx, cmd.UserID, cmd.ConversationPublicID)
	if err != nil {
		return nil, fmt.Errorf("find ai conversation: %w", err)
	}
	if agg == nil {
		return nil, application.NewNotFound("conversation not found")
	}

	profile, err := h.profiles.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("find user profile: %w", err)
	}
	if profile == nil {
		return nil, application.NewNotFound("user profile not found")
	}
	// Lazy daily reset: persist immediately so the fresh window survives even if
	// the AI call below fails.
	if profile.EnsureDailyWindow(h.clock()) {
		if err := h.profiles.Update(ctx, profile); err != nil {
			return nil, fmt.Errorf("reset ai quota window: %w", err)
		}
	}

	// Cheap pre-check so an exhausted user never pays for an AI round-trip. The
	// authoritative check is the atomic decrement after the reply is stored.
	if !profile.HasAIQuotaRemaining() {
		return nil, application.NewQuotaExceeded("AI quota exhausted")
	}

	patientContext, err := h.contextBuilder.Build(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	now := h.clock()
	userTokenCount := estimateTokenCount(content)
	userMessage := &entities.AIMessage{
		ConversationID: agg.Conversation.ID,
		Role:           valueobjects.AIRoleUser,
		Content:        content,
		TokenCount:     &userTokenCount,
		CreatedAt:      now,
	}

	if err := agg.AddMessage(userMessage, now); err != nil {
		return nil, err
	}
	if err := h.messages.Save(ctx, userMessage); err != nil {
		return nil, fmt.Errorf("save user ai message: %w", err)
	}
	if err := h.conversations.Update(ctx, agg); err != nil {
		return nil, fmt.Errorf("touch ai conversation: %w", err)
	}

	chatResp, err := h.ai.ChatWithPatientContext(ctx, dto.PatientContextChatRequest{
		History:             conversationHistory(agg, h.maxContextMessages, h.maxContextTokens),
		PatientContext:      patientContext.Text,
		ConversationSummary: agg.Conversation.Summary,
		Language:            cmd.Language,
	})
	if err != nil {
		return nil, err
	}

	now = h.clock()
	assistantTokenCount := chatResp.OutputTokens
	if assistantTokenCount <= 0 {
		assistantTokenCount = estimateTokenCount(chatResp.Reply)
	}
	assistantMessage := &entities.AIMessage{
		ConversationID: agg.Conversation.ID,
		Role:           valueobjects.AIRoleAssistant,
		Content:        chatResp.Reply,
		TokenCount:     &assistantTokenCount,
		Meta: map[string]any{
			"model":         chatResp.Model,
			"prompt_tokens": chatResp.PromptTokens,
			"output_tokens": chatResp.OutputTokens,
			"context":       patientContext.Meta,
		},
		CreatedAt: now,
	}

	if err := agg.AddMessage(assistantMessage, now); err != nil {
		return nil, err
	}
	if err := h.messages.Save(ctx, assistantMessage); err != nil {
		return nil, fmt.Errorf("save assistant ai message: %w", err)
	}

	// Authoritative, race-free decrement. The pre-check above only avoids paying
	// for an AI call we know will be rejected.
	consumed, err := h.profiles.ConsumeAIQuestion(ctx, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("consume ai question: %w", err)
	}
	if !consumed {
		return nil, application.NewQuotaExceeded("AI quota exhausted")
	}

	// Keep the in-memory copy in step so the quota returned in this response
	// reflects the question just spent.
	if !profile.AIIsPro {
		profile.AIQuestionsUsed++
	}

	if agg.Conversation.Title == nil {
		agg.SetTitle(titleFromMessage(content), now)
	}

	h.applySummaryIfNeeded(ctx, agg, now)

	if err := h.conversations.Update(ctx, agg); err != nil {
		return nil, fmt.Errorf("update ai conversation: %w", err)
	}

	return &dto.SendAIMessageResultDTO{
		Conversation: mapper.AIConversationAggregateToDTO(agg, true),
		Message:      mapper.AIMessageToDTO(userMessage),
		Reply:        mapper.AIMessageToDTO(assistantMessage),
		Model:        chatResp.Model,
		PromptTokens: chatResp.PromptTokens,
		OutputTokens: chatResp.OutputTokens,
		ContextMeta:  patientContext.Meta,
		Quota:        AIQuotaToDTO(profile, now),
	}, nil
}

func (h *SendAIMessageHandler) applySummaryIfNeeded(
	ctx context.Context,
	agg *aggregates.AIConversationAggregate,
	now time.Time,
) {
	if h.summaryTokenThreshold <= 0 || !agg.ShouldSummarise(h.summaryTokenThreshold) {
		return
	}
	last := agg.LastMessage()
	if last == nil || last.ID == 0 {
		return
	}

	summary, err := h.ai.SummarizeChat(ctx, dto.ChatSummarizeRequest{
		Messages: summaryMessages(agg.Messages),
	})
	if err != nil {
		return
	}

	agg.ApplySummary(summary.Summary, last.ID, now)
}

func conversationHistory(agg *aggregates.AIConversationAggregate, maxMessages, maxTokens int) []dto.ChatMessage {
	messages := visibleMessages(agg)
	if maxMessages > 0 && len(messages) > maxMessages {
		messages = messages[len(messages)-maxMessages:]
	}
	messages = limitByTokenBudget(messages, maxTokens)

	for len(messages) > 0 && messages[0].Role != valueobjects.AIRoleUser {
		messages = messages[1:]
	}

	return coalesceChatMessages(messages)
}

func summaryMessages(messages []*entities.AIMessage) []dto.ChatMessage {
	filtered := make([]*entities.AIMessage, 0, len(messages))
	for _, message := range messages {
		if message != nil && message.Role != valueobjects.AIRoleSystem {
			filtered = append(filtered, message)
		}
	}
	return coalesceChatMessages(filtered)
}

func visibleMessages(agg *aggregates.AIConversationAggregate) []*entities.AIMessage {
	messages := make([]*entities.AIMessage, 0, len(agg.Messages))
	for _, message := range agg.Messages {
		if message == nil || message.Role == valueobjects.AIRoleSystem {
			continue
		}
		if agg.Conversation.SummaryUpTo != nil && message.ID > 0 && message.ID <= *agg.Conversation.SummaryUpTo {
			continue
		}
		messages = append(messages, message)
	}
	return messages
}

func limitByTokenBudget(messages []*entities.AIMessage, maxTokens int) []*entities.AIMessage {
	if maxTokens <= 0 || len(messages) == 0 {
		return messages
	}

	total := 0
	window := make([]*entities.AIMessage, 0, len(messages))
	for i := len(messages) - 1; i >= 0; i-- {
		message := messages[i]
		tokens := messageTokenCount(message)
		if len(window) > 0 && total+tokens > maxTokens {
			break
		}
		total += tokens
		window = append([]*entities.AIMessage{message}, window...)
	}
	return window
}

func coalesceChatMessages(messages []*entities.AIMessage) []dto.ChatMessage {
	result := make([]dto.ChatMessage, 0, len(messages))
	for _, message := range messages {
		if message == nil || strings.TrimSpace(message.Content) == "" {
			continue
		}

		role := message.Role.String()
		content := strings.TrimSpace(message.Content)
		last := len(result) - 1
		if last >= 0 && result[last].Role == role {
			result[last].Content += "\n\n" + content
			continue
		}

		result = append(result, dto.ChatMessage{
			Role:    role,
			Content: content,
		})
	}
	return result
}

func messageTokenCount(message *entities.AIMessage) int {
	if message == nil {
		return 0
	}
	if message.TokenCount != nil && *message.TokenCount > 0 {
		return *message.TokenCount
	}
	return estimateTokenCount(message.Content)
}

func estimateTokenCount(text string) int {
	words := len(strings.Fields(text))
	if words == 0 {
		return 1
	}
	return (words * 4 / 3) + 1
}

func titleFromMessage(message string) string {
	title := strings.TrimSpace(message)
	title = strings.ReplaceAll(title, "\n", " ")
	if len(title) > 80 {
		title = strings.TrimSpace(title[:80])
	}
	if title == "" {
		return "New conversation"
	}
	return title
}
