package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

type UpdateFunFactHandler struct {
	funFacts repository.FunFactRepository
}

func NewUpdateFunFactHandler(funFacts repository.FunFactRepository) *UpdateFunFactHandler {
	return &UpdateFunFactHandler{funFacts: funFacts}
}

func (h *UpdateFunFactHandler) Handle(
	ctx context.Context,
	cmd command.UpdateFunFactCommand,
) (struct{}, error) {
	if cmd.ID <= 0 {
		return struct{}{}, errors.New("fun fact id is required")
	}
	if strings.TrimSpace(cmd.Text) == "" {
		return struct{}{}, errors.New("fun fact text is required")
	}
	title := strings.TrimSpace(cmd.Title)
	if title == "" {
		return struct{}{}, errors.New("fun fact title is required")
	}
	if err := validateFunFactRanges(cmd.TargetAgeMin, cmd.TargetAgeMax, cmd.AllergyCategory); err != nil {
		return struct{}{}, err
	}

	existing, err := h.funFacts.FindByID(ctx, cmd.ID)
	if err != nil {
		return struct{}{}, fmt.Errorf("find fun fact: %w", err)
	}
	if existing == nil {
		return struct{}{}, errors.New("fun fact not found")
	}

	existing.Title = title
	existing.Text = strings.TrimSpace(cmd.Text)
	existing.Category = sanitizeOptionalString(cmd.Category)
	existing.TargetCountryCode = sanitizeOptionalString(cmd.TargetCountryCode)
	existing.TargetAgeMin = cmd.TargetAgeMin
	existing.TargetAgeMax = cmd.TargetAgeMax
	existing.AllergyCategory = cmd.AllergyCategory
	if cmd.IsActive != nil {
		existing.IsActive = *cmd.IsActive
	}

	if err := h.funFacts.Update(ctx, existing); err != nil {
		return struct{}{}, fmt.Errorf("update fun fact: %w", err)
	}

	return struct{}{}, nil
}
