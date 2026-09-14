package handlers

import (
	"context"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/mapper"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

type GetFunFactHandler struct {
	funFacts repository.FunFactRepository
}

func NewGetFunFactHandler(funFacts repository.FunFactRepository) *GetFunFactHandler {
	return &GetFunFactHandler{funFacts: funFacts}
}

func (h *GetFunFactHandler) Handle(
	ctx context.Context,
	q query.GetFunFactQuery,
) (*dto.FunFactDTO, error) {
	if q.ID <= 0 {
		return nil, application.NewValidation("fun fact id is required")
	}

	fact, err := h.funFacts.FindByID(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("find fun fact: %w", err)
	}
	if fact == nil {
		return nil, nil
	}

	// is_active is the publication gate: DashboardRepository.getFunFact already
	// filters on it, so a retracted fact must not remain readable here either.
	// Filtered in this query handler rather than in FunFactRepository.FindByID,
	// which the admin update and delete handlers share and which must still
	// return a retracted fact so it can be re-activated.
	if !fact.IsActive {
		return nil, nil
	}

	d := mapper.FunFactToDTO(fact)
	return &d, nil
}
