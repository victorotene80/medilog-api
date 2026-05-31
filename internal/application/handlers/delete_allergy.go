package handlers

import (
	"context"
	"errors"

	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

type DeleteAllergyHandler struct {
	allergyRepository repository.AllergyRepository
}

func NewDeleteAllergyHandler(
	allergyRepository repository.AllergyRepository,
) *DeleteAllergyHandler {
	return &DeleteAllergyHandler{allergyRepository: allergyRepository}
}

func (h *DeleteAllergyHandler) Handle(
	ctx context.Context,
	cmd command.DeleteAllergyCommand,
) (struct{}, error) {
	if cmd.ID <= 0 {
		return struct{}{}, errors.New("allergy id is required")
	}

	existing, err := h.allergyRepository.FindByID(ctx, cmd.ID)
	if err != nil {
		return struct{}{}, err
	}
	if existing == nil {
		return struct{}{}, errors.New("allergy not found")
	}

	if err := h.allergyRepository.Delete(ctx, cmd.ID); err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}
