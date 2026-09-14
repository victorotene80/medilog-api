package handlers

import (
	"context"
	"strings"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type UpdateAllergyHandler struct {
	allergyRepository repository.AllergyRepository
}

func NewUpdateAllergyHandler(
	allergyRepository repository.AllergyRepository,
) *UpdateAllergyHandler {
	return &UpdateAllergyHandler{allergyRepository: allergyRepository}
}

func (h *UpdateAllergyHandler) Handle(
	ctx context.Context,
	cmd command.UpdateAllergyCommand,
) (struct{}, error) {
	if cmd.ID <= 0 {
		return struct{}{}, application.NewValidation("allergy id is required")
	}

	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return struct{}{}, application.NewValidation("allergy name is required")
	}

	category, err := valueobjects.NewAllergyCategory(cmd.Category)
	if err != nil {
		return struct{}{}, err
	}

	existing, err := h.allergyRepository.FindByID(ctx, cmd.ID)
	if err != nil {
		return struct{}{}, err
	}
	if existing == nil {
		return struct{}{}, application.NewNotFound("allergy not found")
	}

	var description *string
	if cmd.Description != nil {
		v := strings.TrimSpace(*cmd.Description)
		if v != "" {
			description = &v
		}
	}

	existing.Name = name
	existing.Category = category
	existing.Description = description

	if err := h.allergyRepository.Update(ctx, existing); err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}
