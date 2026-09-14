package handlers

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

type EmergencyContactHandler struct {
	userRepository      repository.UserAggregateRepository
	emergencyRepository repository.EmergencyContactRepository
	tx                  appContracts.TransactionManager
	clock               func() time.Time
}

func NewEmergencyContactHandler(
	userRepository repository.UserAggregateRepository,
	emergencyRepository repository.EmergencyContactRepository,
	tx appContracts.TransactionManager,
	clock func() time.Time,
) *EmergencyContactHandler {
	return &EmergencyContactHandler{
		userRepository:      userRepository,
		emergencyRepository: emergencyRepository,
		tx:                  tx,
		clock:               clock,
	}
}

func (h *EmergencyContactHandler) Handle(
	ctx context.Context,
	cmd command.EmergencyContactCommand,
) (*dto.EmergencyContactDTO, error) {
	if cmd.UserID <= 0 {
		return nil, application.NewValidation("user id is required")
	}

	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, application.NewValidation("emergency contact name is required")
	}

	relationship := strings.TrimSpace(cmd.Relationship)
	if relationship == "" {
		return nil, application.NewValidation("emergency contact relationship is required")
	}

	phone := strings.TrimSpace(cmd.Phone)
	if phone == "" {
		return nil, application.NewValidation("emergency contact phone is required")
	}

	existingContact, err := h.emergencyRepository.FindByUserIDAndPhone(ctx, cmd.UserID, phone)
	if err != nil {
		return nil, err
	}

	if existingContact != nil {
		return nil, application.NewConflict("emergency contact with this phone number already exists")
	}

	now := h.clock().UTC()

	existing, err := h.emergencyRepository.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	// The first contact is always primary. Without this a user can end up with
	// no primary at all, which leaves GetUserHandler falling back to "first by
	// id" forever and would break the one-primary-per-user invariant.
	isPrimary := cmd.IsPrimary || len(existing) == 0

	emergencyContact := entities.EmergencyContact{
		UserID:       cmd.UserID,
		Name:         name,
		Relationship: relationship,
		Phone:        phone,
		CountryCode:  cmd.CountryCode,
		// Always inserted non-primary, even when this contact is destined to be
		// primary. Inserting it as primary while the previous primary is still
		// flagged would momentarily give the user two, which the
		// one-primary-per-user unique index rejects outright. SetPrimary below
		// demotes and promotes in a single transaction instead.
		IsPrimary: false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// One unit of work. Creating the contact is what completes onboarding, and
	// the two used to commit on separate connections: if the user update failed
	// after the contact was saved, the retry hit the duplicate-phone conflict
	// above while RequireOnboardingCompleted kept the account locked out of the
	// whole app — with no way back except a different phone number.
	if err := h.tx.Do(ctx, func(ctx context.Context) error {
		if err := h.emergencyRepository.Save(ctx, &emergencyContact); err != nil {
			return err
		}

		if isPrimary {
			if err := h.emergencyRepository.SetPrimary(ctx, cmd.UserID, emergencyContact.ID, now); err != nil {
				return err
			}
			emergencyContact.IsPrimary = true
		}

		agg, err := h.userRepository.FindByID(ctx, cmd.UserID)
		if err != nil {
			return err
		}

		if agg == nil {
			return application.NewNotFound("user not found")
		}

		// Adding a second contact must not fail, so the already-completed case
		// is a no-op rather than the error CompleteOnboarding returns.
		if agg.User.IsOnboardingCompleted {
			return nil
		}

		// Through the aggregate rather than by assigning the field: this is the
		// transition CompleteOnboarding exists for, and it is what raises
		// user.onboarding_completed.
		if err := agg.CompleteOnboarding(now); err != nil {
			return err
		}

		return h.userRepository.Update(ctx, agg)
	}); err != nil {
		return nil, err
	}

	return EmergencyContactToDTO(&emergencyContact), nil
}

// EmergencyContactToDTO maps an emergency contact for the application boundary.
// ID is the public UUID: the internal row id must never leave the server, and
// every emergency-contact route addresses contacts by public id.
func EmergencyContactToDTO(c *entities.EmergencyContact) *dto.EmergencyContactDTO {
	if c == nil {
		return nil
	}

	return &dto.EmergencyContactDTO{
		ID:           c.PublicID,
		UserID:       strconv.FormatInt(c.UserID, 10),
		Name:         c.Name,
		Relationship: c.Relationship,
		Phone:        c.Phone,
		CountryCode:  c.CountryCode,
		IsPrimary:    c.IsPrimary,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}
