package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

type CompleteMedicationHandler struct {
	medications domainRepo.MedicationRepository
}

func NewCompleteMedicationHandler(medications domainRepo.MedicationRepository) *CompleteMedicationHandler {
	return &CompleteMedicationHandler{medications: medications}
}

func (h *CompleteMedicationHandler) Handle(ctx context.Context, cmd command.CompleteMedicationCommand) (struct{}, error) {
	agg, err := h.medications.FindByPublicID(ctx, cmd.UserID, cmd.PublicID)
	if err != nil {
		return struct{}{}, fmt.Errorf("find medication: %w", err)
	}
	if agg == nil {
		return struct{}{}, application.NewNotFound("medication not found")
	}

	if err := agg.Complete(time.Now().UTC()); err != nil {
		return struct{}{}, err
	}

	if err := h.medications.Update(ctx, agg); err != nil {
		return struct{}{}, fmt.Errorf("update medication: %w", err)
	}

	return struct{}{}, nil
}
