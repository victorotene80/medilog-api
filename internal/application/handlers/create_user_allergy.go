package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type CreateUserAllergiesHandler struct {
	userAllergyRepository repository.UserAllergyRepository
	allergyRepository     repository.AllergyRepository
	clock                 func() time.Time
}

func NewCreateUserAllergiesHandler(
	userAllergyRepository repository.UserAllergyRepository,
	allergyRepository repository.AllergyRepository,
	clock func() time.Time,
) *CreateUserAllergiesHandler {
	if clock == nil {
		clock = func() time.Time {
			return time.Now().UTC()
		}
	}

	return &CreateUserAllergiesHandler{
		userAllergyRepository: userAllergyRepository,
		allergyRepository:     allergyRepository,
		clock:                 clock,
	}
}

func (h *CreateUserAllergiesHandler) Handle(
	ctx context.Context,
	cmd command.CreateUserAllergiesCommand,
) (struct{}, error) {
	if cmd.UserID <= 0 {
		return struct{}{}, application.NewValidation("user id is required")
	}

	if len(cmd.Allergies) == 0 {
		return struct{}{}, application.NewValidation("at least one allergy is required")
	}

	now := h.clock().UTC()

	seenAllergyIDs := make(map[int64]struct{})

	// One dedupe set, keyed on the *resolved* name. Keying by input shape
	// instead — one set for ids, another for free-text names — let
	// [{allergy_id: 5}, {name: "peanut"}] through when allergy 5 is "Peanut",
	// because neither set saw the other. The pair then collided on
	// ux_user_allergies_user_name inside SaveBatch and surfaced as a 500
	// carrying the index name, table name and driver text.
	seenNames := make(map[string]struct{})

	userAllergies := make([]*entities.UserAllergy, 0, len(cmd.Allergies))

	for _, item := range cmd.Allergies {
		if item.AllergyID != nil {
			if *item.AllergyID <= 0 {
				return struct{}{}, application.NewValidation("invalid allergy id")
			}

			if _, exists := seenAllergyIDs[*item.AllergyID]; exists {
				return struct{}{}, application.NewConflict("duplicate allergy selected in request")
			}

			seenAllergyIDs[*item.AllergyID] = struct{}{}
		} else if strings.TrimSpace(item.Name) == "" {
			return struct{}{}, application.NewValidation("allergy name is required when allergy id is not provided")
		}

		userAllergy, err := h.buildUserAllergy(ctx, cmd.UserID, item, now)
		if err != nil {
			return struct{}{}, err
		}

		// Checked after buildUserAllergy because that is where a catalogue item's
		// name is rewritten to the master record's — the rewrite is what makes
		// the two input shapes the same row.
		nameKey := strings.ToLower(strings.TrimSpace(userAllergy.Name))
		if _, exists := seenNames[nameKey]; exists {
			return struct{}{}, application.NewConflict("duplicate allergy in request")
		}

		seenNames[nameKey] = struct{}{}

		userAllergies = append(userAllergies, userAllergy)
	}

	if err := h.userAllergyRepository.SaveBatch(ctx, userAllergies); err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}

func (h *CreateUserAllergiesHandler) buildUserAllergy(
	ctx context.Context,
	userID int64,
	item command.UserAllergyItemCommand,
	now time.Time,
) (*entities.UserAllergy, error) {
	name := strings.TrimSpace(item.Name)
	description := normalizeOptionalString(item.Description)

	if item.AllergyID == nil && name == "" {
		return nil, application.NewValidation("allergy name is required when allergy id is not provided")
	}

	if item.AllergyID != nil && *item.AllergyID <= 0 {
		return nil, application.NewValidation("invalid allergy id")
	}

	var category valueobjects.AllergyCategory

	if item.AllergyID != nil {
		masterAllergy, findErr := h.allergyRepository.FindByID(ctx, *item.AllergyID)
		if findErr != nil {
			return nil, findErr
		}

		if masterAllergy == nil {
			return nil, application.NewNotFound("selected allergy does not exist")
		}

		existingUserAllergy, err := h.userAllergyRepository.FindByUserIDAndAllergyID(
			ctx,
			userID,
			*item.AllergyID,
		)
		if err != nil {
			return nil, err
		}

		if existingUserAllergy != nil {
			return nil, application.NewConflict("user already has this allergy")
		}

		name = masterAllergy.Name
		description = masterAllergy.Description
		category = masterAllergy.Category
	} else {
		existingUserAllergy, err := h.userAllergyRepository.FindByUserIDAndName(ctx, userID, name)
		if err != nil {
			return nil, err
		}

		if existingUserAllergy != nil {
			return nil, application.NewConflict("user already has an allergy with this name")
		}

		parsedCategory, categoryErr := valueobjects.NewAllergyCategory(item.Category)
		if categoryErr != nil {
			return nil, application.NewValidation("invalid allergy category")
		}

		category = parsedCategory
	}

	var severity *valueobjects.AllergySeverity
	if item.Severity != nil {
		parsedSeverity, severityErr := valueobjects.NewAllergySeverity(*item.Severity)
		if severityErr != nil {
			return nil, application.NewValidation("invalid allergy severity")
		}

		severity = &parsedSeverity
	}

	return &entities.UserAllergy{
		PublicID:    uuid.NewString(),
		UserID:      userID,
		AllergyID:   item.AllergyID,
		Name:        name,
		Description: description,
		Severity:    severity,
		Category:    category,
		CreatedAt:   now,
	}, nil
}
func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	v := strings.TrimSpace(*value)
	if v == "" {
		return nil
	}

	return &v
}
