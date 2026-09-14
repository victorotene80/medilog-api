package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"
)

// UpdateUserHandler applies a partial update to the caller's own profile.
//
// It writes across two tables (users and user_profiles) but goes through
// UserAggregateRepository.Update, which already persists both inside one
// transaction — so a bad value in the profile half cannot leave the user half
// half-written.
type UpdateUserHandler struct {
	userRepo repository.UserAggregateRepository
	clock    func() time.Time
}

func NewUpdateUserHandler(
	userRepo repository.UserAggregateRepository,
	clock func() time.Time,
) *UpdateUserHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &UpdateUserHandler{userRepo: userRepo, clock: clock}
}

func (h *UpdateUserHandler) Handle(
	ctx context.Context,
	cmd command.UpdateUserCommand,
) (*dto.GetUserDTO, error) {
	if cmd.UserID <= 0 {
		return nil, application.NewValidation("user id is required")
	}

	agg, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	if agg == nil || agg.User == nil {
		return nil, application.NewNotFound("user not found")
	}

	// Everything is validated before anything is mutated, so a rejected field
	// never leaves the aggregate partially updated.
	if err := applyUserFields(agg.User, cmd); err != nil {
		return nil, err
	}

	if err := applyProfileFields(agg, cmd); err != nil {
		return nil, err
	}

	agg.User.UpdatedAt = h.clock().UTC()

	if err := h.userRepo.Update(ctx, agg); err != nil {
		return nil, err
	}

	return UserAggregateToGetUserDTO(agg), nil
}

func applyUserFields(user *entities.User, cmd command.UpdateUserCommand) error {
	if cmd.FirstName != nil {
		name := strings.TrimSpace(*cmd.FirstName)
		if name == "" {
			return application.NewValidation("first name cannot be empty")
		}
		user.FirstName = name
	}

	if cmd.LastName != nil {
		name := strings.TrimSpace(*cmd.LastName)
		if name == "" {
			return application.NewValidation("last name cannot be empty")
		}
		user.LastName = name
	}

	if cmd.DOB != nil {
		user.DateOfBirth = cmd.DOB
	}

	if cmd.Sex != nil {
		sex, err := valueobjects.NewSex(*cmd.Sex)
		if err != nil {
			return application.NewValidation(err.Error())
		}
		user.Sex = &sex
	}

	if cmd.BloodType != nil {
		bloodType, err := valueobjects.NewBloodType(strings.ToUpper(strings.TrimSpace(*cmd.BloodType)))
		if err != nil {
			return application.NewValidation(err.Error())
		}
		user.BloodType = &bloodType
	}

	if cmd.AvatarURL != nil {
		avatar := strings.TrimSpace(*cmd.AvatarURL)
		user.AvatarURL = &avatar
	}

	if cmd.CountryCode != nil {
		countryCode, err := valueobjects.NewCountryCode(*cmd.CountryCode)
		if err != nil {
			return application.NewValidation(err.Error())
		}
		user.CountryCode = &countryCode
	}

	return nil
}

func applyProfileFields(agg *aggregates.UserAggregate, cmd command.UpdateUserCommand) error {
	if cmd.Height == nil && cmd.Weight == nil &&
		cmd.WeightUnit == nil && cmd.TemperatureUnit == nil {
		return nil
	}

	// A user who registered before profiles existed, or whose profile row was
	// never created, still needs these fields to land somewhere.
	if agg.Profile == nil {
		agg.Profile = entities.NewDefaultUserProfile(agg.User.ID)
	}

	if cmd.Height != nil {
		if *cmd.Height <= 0 {
			return application.NewValidation("height must be greater than zero")
		}
		agg.Profile.Height = cmd.Height
	}

	if cmd.Weight != nil {
		if *cmd.Weight <= 0 {
			return application.NewValidation("weight must be greater than zero")
		}
		agg.Profile.Weight = cmd.Weight
	}

	if cmd.WeightUnit != nil {
		unit := strings.ToLower(strings.TrimSpace(*cmd.WeightUnit))
		if unit != "kg" && unit != "lb" {
			return application.NewValidation("weight unit must be kg or lb")
		}
		agg.Profile.WeightUnit = unit
	}

	if cmd.TemperatureUnit != nil {
		unit := strings.ToLower(strings.TrimSpace(*cmd.TemperatureUnit))
		if unit != "celsius" && unit != "fahrenheit" {
			return application.NewValidation("temperature unit must be celsius or fahrenheit")
		}
		agg.Profile.TemperatureUnit = unit
	}

	return nil
}
