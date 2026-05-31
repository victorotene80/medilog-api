package mapper

import (
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

func AIConversationAggregateToDTO(agg *aggregates.AIConversationAggregate, includeMessages bool) dto.AIConversationDTO {
	if agg == nil || agg.Conversation == nil {
		return dto.AIConversationDTO{}
	}

	conversation := agg.Conversation
	result := dto.AIConversationDTO{
		ID:                  conversation.ID,
		PublicID:            conversation.PublicID,
		UserID:              conversation.UserID,
		Title:               conversation.Title,
		RelatedMedicationID: conversation.RelatedMedicationID,
		RelatedVisitID:      conversation.RelatedVisitID,
		Summary:             conversation.Summary,
		Status:              conversation.Status.String(),
		LastMessageAt:       conversation.LastMessageAt,
		CreatedAt:           conversation.CreatedAt,
		UpdatedAt:           conversation.UpdatedAt,
	}

	if includeMessages {
		result.Messages = AIMessagesToDTO(agg.Messages)
	}

	return result
}

func AIConversationAggregatesToDTO(aggs []*aggregates.AIConversationAggregate, includeMessages bool) []dto.AIConversationDTO {
	result := make([]dto.AIConversationDTO, 0, len(aggs))
	for _, agg := range aggs {
		result = append(result, AIConversationAggregateToDTO(agg, includeMessages))
	}
	return result
}

func AIMessageToDTO(message *entities.AIMessage) dto.AIMessageDTO {
	if message == nil {
		return dto.AIMessageDTO{}
	}

	return dto.AIMessageDTO{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		Role:           message.Role.String(),
		Content:        message.Content,
		TokenCount:     message.TokenCount,
		Meta:           message.Meta,
		CreatedAt:      message.CreatedAt,
	}
}

func AIMessagesToDTO(messages []*entities.AIMessage) []dto.AIMessageDTO {
	result := make([]dto.AIMessageDTO, 0, len(messages))
	for _, message := range messages {
		result = append(result, AIMessageToDTO(message))
	}
	return result
}
