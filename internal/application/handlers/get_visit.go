package handlers

import (
	"context"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/mapper"
	"github.com/victorotene80/medilog-api/internal/application/query"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

type GetVisitHandler struct {
	visits domainRepo.VisitRepository
}

func NewGetVisitHandler(visits domainRepo.VisitRepository) *GetVisitHandler {
	return &GetVisitHandler{visits: visits}
}

func (h *GetVisitHandler) Handle(ctx context.Context, q query.GetVisitQuery) (*dto.VisitDTO, error) {
	visit, err := h.visits.FindByPublicID(ctx, q.UserID, q.PublicID)
	if err != nil {
		return nil, fmt.Errorf("find visit: %w", err)
	}
	if visit == nil {
		return nil, nil
	}

	result := mapper.VisitToDTO(visit)
	return &result, nil
}
