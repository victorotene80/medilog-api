package handlers

import (
	"context"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/mapper"
	"github.com/victorotene80/medilog-api/internal/application/query"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

type GetMedicationHandler struct {
	medications domainRepo.MedicationRepository
}

func NewGetMedicationHandler(medications domainRepo.MedicationRepository) *GetMedicationHandler {
	return &GetMedicationHandler{medications: medications}
}

func (h *GetMedicationHandler) Handle(ctx context.Context, q query.GetMedicationQuery) (*dto.MedicationDTO, error) {
	agg, err := h.medications.FindByPublicID(ctx, q.UserID, q.PublicID)
	if err != nil {
		return nil, fmt.Errorf("find medication: %w", err)
	}
	if agg == nil {
		return nil, application.NewNotFound("medication not found")
	}

	result := mapper.MedicationAggregateToDTO(agg)
	return &result, nil
}
