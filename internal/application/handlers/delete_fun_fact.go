package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

type DeleteFunFactHandler struct {
	funFacts repository.FunFactRepository
}

func NewDeleteFunFactHandler(funFacts repository.FunFactRepository) *DeleteFunFactHandler {
	return &DeleteFunFactHandler{funFacts: funFacts}
}

func (h *DeleteFunFactHandler) Handle(
	ctx context.Context,
	cmd command.DeleteFunFactCommand,
) (struct{}, error) {
	if cmd.ID <= 0 {
		return struct{}{}, errors.New("fun fact id is required")
	}

	existing, err := h.funFacts.FindByID(ctx, cmd.ID)
	if err != nil {
		return struct{}{}, fmt.Errorf("find fun fact: %w", err)
	}
	if existing == nil {
		return struct{}{}, errors.New("fun fact not found")
	}

	if err := h.funFacts.Delete(ctx, cmd.ID); err != nil {
		return struct{}{}, fmt.Errorf("delete fun fact: %w", err)
	}

	return struct{}{}, nil
}
