package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"
)

// UpdateUserAllergyHandler edits one of the user's allergy records in place.
//
// Before this existed the client had to delete and re-add to make an edit,
// which changed the record's public id on every save.
type UpdateUserAllergyHandler struct {
	userAllergyRepository repository.UserAllergyRepository
	clock                 func() time.Time
}

func NewUpdateUserAllergyHandler(
	userAllergyRepository repository.UserAllergyRepository,
	clock func() time.Time,
) *UpdateUserAllergyHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &UpdateUserAllergyHandler{
		userAllergyRepository: userAllergyRepository,
		clock:                 clock,
	}
}

func (h *UpdateUserAllergyHandler) Handle(
	ctx context.Context,
	cmd command.UpdateUserAllergyCommand,
) (*dto.UserAllergyDTO, error) {
	if cmd.UserID <= 0 {
		return nil, application.NewValidation("user id is required")
	}

	publicID := strings.TrimSpace(cmd.PublicID)
	if publicID == "" {
		return nil, application.NewValidation("allergy public id is required")
	}

	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, application.NewValidation("allergy name is required")
	}

	category, err := valueobjects.NewAllergyCategory(cmd.Category)
	if err != nil {
		return nil, application.NewValidation("invalid allergy category")
	}

	var severity *valueobjects.AllergySeverity
	if cmd.Severity != nil {
		parsed, sevErr := valueobjects.NewAllergySeverity(*cmd.Severity)
		if sevErr != nil {
			return nil, application.NewValidation("invalid allergy severity")
		}
		severity = &parsed
	}

	allergy, err := h.userAllergyRepository.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, err
	}

	// FindByPublicID is not user-scoped, so ownership is enforced here. Another
	// user's record is reported as missing rather than forbidden.
	if allergy == nil || allergy.UserID != cmd.UserID {
		return nil, application.NewNotFound("allergy not found")
	}

	renamed := !strings.EqualFold(name, allergy.Name)

	// Renaming onto one of the user's other allergies collides with the
	// (user_id, lower(name)) uniqueness rule.
	if renamed {
		duplicate, dupErr := h.userAllergyRepository.FindByUserIDAndName(ctx, cmd.UserID, name)
		if dupErr != nil {
			return nil, dupErr
		}

		if duplicate != nil && duplicate.ID != allergy.ID {
			return nil, application.NewConflict("user already has an allergy with this name")
		}
	}

	allergy.Name = name
	allergy.Description = normalizeOptionalString(cmd.Description)
	allergy.Severity = severity
	allergy.Category = category
	allergy.UpdatedAt = h.clock().UTC()

	// Renaming detaches the record from the reference catalogue entry it was
	// created from, since it no longer describes that catalogue row.
	if renamed {
		allergy.AllergyID = nil
	}

	if err := h.userAllergyRepository.Update(ctx, allergy); err != nil {
		return nil, err
	}

	result := userAllergyToDTO(allergy)

	return &result, nil
}
