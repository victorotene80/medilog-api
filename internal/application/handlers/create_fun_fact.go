package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type CreateFunFactHandler struct {
	funFacts repository.FunFactRepository
}

func NewCreateFunFactHandler(funFacts repository.FunFactRepository) *CreateFunFactHandler {
	return &CreateFunFactHandler{funFacts: funFacts}
}

func (h *CreateFunFactHandler) Handle(
	ctx context.Context,
	cmd command.CreateFunFactCommand,
) (struct{}, error) {
	if strings.TrimSpace(cmd.Text) == "" {
		return struct{}{}, application.NewValidation("fun fact text is required")
	}
	title := strings.TrimSpace(cmd.Title)
	if title == "" {
		return struct{}{}, application.NewValidation("fun fact title is required")
	}

	if err := validateFunFactRanges(cmd.TargetAgeMin, cmd.TargetAgeMax, cmd.AllergyCategory); err != nil {
		return struct{}{}, err
	}

	now := time.Now().UTC()
	fact := &entities.FunFact{
		Title:             title,
		Text:              strings.TrimSpace(cmd.Text),
		Category:          sanitizeOptionalString(cmd.Category),
		TargetCountryCode: sanitizeOptionalString(cmd.TargetCountryCode),
		TargetAgeMin:      cmd.TargetAgeMin,
		TargetAgeMax:      cmd.TargetAgeMax,
		AllergyCategory:   cmd.AllergyCategory,
		IsActive:          true,
		CreatedAt:         now,
	}

	if cmd.IsActive != nil {
		fact.IsActive = *cmd.IsActive
	}

	if err := h.funFacts.Save(ctx, fact); err != nil {
		return struct{}{}, fmt.Errorf("save fun fact: %w", err)
	}

	return struct{}{}, nil
}

func validateFunFactRanges(
	minAge *int,
	maxAge *int,
	allergyCategory *int,
) error {
	if minAge != nil && *minAge < 0 {
		return application.NewValidation("target_age_min cannot be negative")
	}

	if maxAge != nil && *maxAge < 0 {
		return application.NewValidation("target_age_max cannot be negative")
	}

	if minAge != nil && maxAge != nil && *minAge > *maxAge {
		return application.NewValidation("target_age_min cannot be greater than target_age_max")
	}

	if allergyCategory != nil {
		if _, err := valueobjects.NewAllergyCategory(*allergyCategory); err != nil {
			return fmt.Errorf("invalid allergy_category: %w", err)
		}
	}

	return nil
}

func sanitizeOptionalString(v *string) *string {
	if v == nil {
		return nil
	}

	s := strings.TrimSpace(*v)
	if s == "" {
		return nil
	}

	return &s
}
