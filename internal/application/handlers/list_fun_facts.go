package handlers

import (
	"context"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/mapper"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

type ListFunFactsHandler struct {
	funFacts repository.FunFactRepository
}

func NewListFunFactsHandler(funFacts repository.FunFactRepository) *ListFunFactsHandler {
	return &ListFunFactsHandler{funFacts: funFacts}
}

func (h *ListFunFactsHandler) Handle(
	ctx context.Context,
	q query.ListFunFactsQuery,
) ([]dto.FunFactDTO, error) {
	facts, err := h.funFacts.FindAll(ctx, q.ActiveOnly)
	if err != nil {
		return nil, fmt.Errorf("list fun facts: %w", err)
	}

	result := make([]dto.FunFactDTO, 0, len(facts))
	for _, fact := range facts {
		result = append(result, mapper.FunFactToDTO(fact))
	}

	return result, nil
}
