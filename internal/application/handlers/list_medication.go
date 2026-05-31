package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/mapper"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

type ListMedicationsHandler struct {
	medications domainRepo.MedicationRepository
}

func NewListMedicationsHandler(medications domainRepo.MedicationRepository) *ListMedicationsHandler {
	return &ListMedicationsHandler{medications: medications}
}

func (h *ListMedicationsHandler) Handle(ctx context.Context, q query.ListMedicationsQuery) ([]dto.MedicationDTO, error) {
	var (
		aggs []*aggregates.MedicationAggregate
		err  error
	)

	if q.ActiveOnly {
		aggs, err = h.medications.FindActiveByUserID(ctx, q.UserID, time.Now().UTC())
	} else {
		aggs, err = h.medications.FindByUserID(ctx, q.UserID)
	}
	if err != nil {
		return nil, fmt.Errorf("list medications: %w", err)
	}

	result := make([]dto.MedicationDTO, 0, len(aggs))
	for _, agg := range aggs {
		result = append(result, mapper.MedicationAggregateToDTO(agg))
	}
	return result, nil
}
