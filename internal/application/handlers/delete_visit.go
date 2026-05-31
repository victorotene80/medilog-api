package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application/command"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

type DeleteVisitHandler struct {
	visits domainRepo.VisitRepository
}

func NewDeleteVisitHandler(visits domainRepo.VisitRepository) *DeleteVisitHandler {
	return &DeleteVisitHandler{visits: visits}
}

func (h *DeleteVisitHandler) Handle(ctx context.Context, cmd command.DeleteVisitCommand) (struct{}, error) {
	visit, err := h.visits.FindByPublicID(ctx, cmd.UserID, cmd.PublicID)
	if err != nil {
		return struct{}{}, fmt.Errorf("find visit: %w", err)
	}
	if visit == nil {
		return struct{}{}, errors.New("visit not found")
	}

	if err := h.visits.Delete(ctx, visit.ID); err != nil {
		return struct{}{}, fmt.Errorf("delete visit: %w", err)
	}

	return struct{}{}, nil
}
