package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application/command"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

type DeleteMedicationHandler struct {
	medications domainRepo.MedicationRepository
}

func NewDeleteMedicationHandler(medications domainRepo.MedicationRepository) *DeleteMedicationHandler {
	return &DeleteMedicationHandler{medications: medications}
}

func (h *DeleteMedicationHandler) Handle(ctx context.Context, cmd command.DeleteMedicationCommand) (struct{}, error) {
	agg, err := h.medications.FindByPublicID(ctx, cmd.UserID, cmd.PublicID)
	if err != nil {
		return struct{}{}, fmt.Errorf("find medication: %w", err)
	}
	if agg == nil {
		return struct{}{}, errors.New("medication not found")
	}

	if err := h.medications.Delete(ctx, agg.Medication.ID); err != nil {
		return struct{}{}, fmt.Errorf("delete medication: %w", err)
	}

	return struct{}{}, nil
}
