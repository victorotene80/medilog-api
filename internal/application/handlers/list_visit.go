package handlers

import (
	"context"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/mapper"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

type ListVisitsHandler struct {
	visits domainRepo.VisitRepository
}

func NewListVisitsHandler(visits domainRepo.VisitRepository) *ListVisitsHandler {
	return &ListVisitsHandler{visits: visits}
}

func (h *ListVisitsHandler) Handle(ctx context.Context, q query.ListVisitsQuery) ([]dto.VisitDTO, error) {
	results, err := h.findVisits(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list visits: %w", err)
	}

	return mapper.VisitsToDTO(results), nil
}

func (h *ListVisitsHandler) findVisits(ctx context.Context, q query.ListVisitsQuery) ([]*entities.Visit, error) {
	if q.From != nil && q.To != nil {
		return h.visits.FindByUserIDAndDateRange(ctx, q.UserID, *q.From, *q.To)
	}
	return h.visits.FindByUserID(ctx, q.UserID)
}
