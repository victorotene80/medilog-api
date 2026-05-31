package handlers

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type AllergyCreationHandler struct {
	allergyRepository repository.AllergyRepository
	clock             func() time.Time
}

func NewAllergyCreationHandler(
	allergyRepository repository.AllergyRepository,
	eventPublisher appContracts.MessagePublisher,
	clock func() time.Time,
) *AllergyCreationHandler {
	return &AllergyCreationHandler{
		allergyRepository: allergyRepository,
		clock:             clock,
	}
}

func (h *AllergyCreationHandler) Handle(
	ctx context.Context,
	cmd command.AllergyCommand,
) (struct{}, error) {
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return struct{}{}, errors.New("allergy name is required")
	}

	category, err := valueobjects.NewAllergyCategory(cmd.Category)
	if err != nil {
		return struct{}{}, err
	}

	var description *string
	if cmd.Description != nil {
		v := strings.TrimSpace(*cmd.Description)
		if v != "" {
			description = &v
		}
	}

	allergy := entities.Allergy{
		Name:        name,
		Category:    category,
		Description: description,
		CreatedAt:   h.clock().UTC(),
	}

	if err := h.allergyRepository.Save(ctx, &allergy); err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}
