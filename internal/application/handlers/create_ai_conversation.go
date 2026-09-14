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
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type CreateAIConversationHandler struct {
	conversations domainRepo.AIConversationRepository
	medications   domainRepo.MedicationRepository
	visits        domainRepo.VisitRepository
	clock         func() time.Time
}

func NewCreateAIConversationHandler(
	conversations domainRepo.AIConversationRepository,
	medications domainRepo.MedicationRepository,
	visits domainRepo.VisitRepository,
	clock func() time.Time,
) *CreateAIConversationHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}

	return &CreateAIConversationHandler{
		conversations: conversations,
		medications:   medications,
		visits:        visits,
		clock:         clock,
	}
}

func (h *CreateAIConversationHandler) Handle(
	ctx context.Context,
	cmd command.CreateAIConversationCommand,
) (dto.AIConversationDTO, error) {
	if cmd.UserID <= 0 {
		return dto.AIConversationDTO{}, application.NewValidation("user id is required")
	}

	relatedMedicationID, err := h.resolveMedicationID(ctx, cmd.UserID, cmd.RelatedMedicationPublicID)
	if err != nil {
		return dto.AIConversationDTO{}, err
	}

	relatedVisitID, err := h.resolveVisitID(ctx, cmd.UserID, cmd.RelatedVisitPublicID)
	if err != nil {
		return dto.AIConversationDTO{}, err
	}

	now := h.clock()
	title := cleanOptionalString(cmd.Title)
	conversation := &entities.AIConversation{
		UserID:              cmd.UserID,
		Title:               title,
		RelatedMedicationID: relatedMedicationID,
		RelatedVisitID:      relatedVisitID,
		Status:              valueobjects.ConversationStatusActive,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	agg := aggregates.NewAIConversationAggregate(conversation)
	if err := h.conversations.Save(ctx, agg); err != nil {
		return dto.AIConversationDTO{}, fmt.Errorf("save ai conversation: %w", err)
	}

	return mapper.AIConversationAggregateToDTO(agg, true), nil
}

func (h *CreateAIConversationHandler) resolveMedicationID(
	ctx context.Context,
	userID int64,
	publicID *string,
) (*int64, error) {
	if publicID == nil || strings.TrimSpace(*publicID) == "" {
		return nil, nil
	}
	if h.medications == nil {
		return nil, errors.New("medication repository is not configured")
	}

	agg, err := h.medications.FindByPublicID(ctx, userID, strings.TrimSpace(*publicID))
	if err != nil {
		return nil, fmt.Errorf("find related medication: %w", err)
	}
	if agg == nil || agg.Medication == nil {
		return nil, application.NewNotFound("related medication not found")
	}

	id := agg.Medication.ID
	return &id, nil
}

func (h *CreateAIConversationHandler) resolveVisitID(
	ctx context.Context,
	userID int64,
	publicID *string,
) (*int64, error) {
	if publicID == nil || strings.TrimSpace(*publicID) == "" {
		return nil, nil
	}
	if h.visits == nil {
		return nil, errors.New("visit repository is not configured")
	}

	visit, err := h.visits.FindByPublicID(ctx, userID, strings.TrimSpace(*publicID))
	if err != nil {
		return nil, fmt.Errorf("find related visit: %w", err)
	}
	if visit == nil {
		return nil, application.NewNotFound("related visit not found")
	}

	id := visit.ID
	return &id, nil
}

func cleanOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
