package handlers

import (
	"context"
	"strings"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

type DeleteUserAllergyHandler struct {
	userAllergyRepository repository.UserAllergyRepository
}

func NewDeleteUserAllergyHandler(
	userAllergyRepository repository.UserAllergyRepository,
) *DeleteUserAllergyHandler {
	return &DeleteUserAllergyHandler{userAllergyRepository: userAllergyRepository}
}

func (h *DeleteUserAllergyHandler) Handle(
	ctx context.Context,
	cmd command.DeleteUserAllergyCommand,
) (struct{}, error) {
	if cmd.UserID <= 0 {
		return struct{}{}, application.NewValidation("user id is required")
	}
	if strings.TrimSpace(cmd.PublicID) == "" {
		return struct{}{}, application.NewValidation("allergy public id is required")
	}

	if err := h.userAllergyRepository.DeleteByPublicID(ctx, cmd.UserID, cmd.PublicID); err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}
