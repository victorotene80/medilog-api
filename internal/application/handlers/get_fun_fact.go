package handlers

import (
	"context"
	"errors"
	"fmt"

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
		return nil, errors.New("fun fact id is required")
	}

	fact, err := h.funFacts.FindByID(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("find fun fact: %w", err)
	}
	if fact == nil {
		return nil, nil
	}

	d := mapper.FunFactToDTO(fact)
	return &d, nil
}
