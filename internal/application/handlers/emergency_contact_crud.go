package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"
)

// ListEmergencyContactsHandler returns every contact belonging to the caller.
type ListEmergencyContactsHandler struct {
	emergencyRepository repository.EmergencyContactRepository
}

func NewListEmergencyContactsHandler(
	emergencyRepository repository.EmergencyContactRepository,
) *ListEmergencyContactsHandler {
	return &ListEmergencyContactsHandler{emergencyRepository: emergencyRepository}
}

func (h *ListEmergencyContactsHandler) Handle(
	ctx context.Context,
	q query.ListEmergencyContactsQuery,
) ([]*dto.EmergencyContactDTO, error) {
	if q.UserID <= 0 {
		return nil, application.NewValidation("user id is required")
	}

	contacts, err := h.emergencyRepository.FindByUserID(ctx, q.UserID)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.EmergencyContactDTO, 0, len(contacts))
	for _, c := range contacts {
		result = append(result, EmergencyContactToDTO(c))
	}

	return result, nil
}

// UpdateEmergencyContactHandler fully replaces a contact.
type UpdateEmergencyContactHandler struct {
	emergencyRepository repository.EmergencyContactRepository
	clock               func() time.Time
}

func NewUpdateEmergencyContactHandler(
	emergencyRepository repository.EmergencyContactRepository,
	clock func() time.Time,
) *UpdateEmergencyContactHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &UpdateEmergencyContactHandler{
		emergencyRepository: emergencyRepository,
		clock:               clock,
	}
}

func (h *UpdateEmergencyContactHandler) Handle(
	ctx context.Context,
	cmd command.UpdateEmergencyContactCommand,
) (*dto.EmergencyContactDTO, error) {
	if cmd.UserID <= 0 {
		return nil, application.NewValidation("user id is required")
	}

	publicID := strings.TrimSpace(cmd.PublicID)
	if publicID == "" {
		return nil, application.NewValidation("emergency contact id is required")
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

	contact, err := h.emergencyRepository.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, err
	}

	// Ownership is checked here rather than in the query because FindByPublicID
	// is not user-scoped. Report someone else's contact as missing, not
	// forbidden, so the response cannot confirm the id exists.
	if contact == nil || contact.UserID != cmd.UserID {
		return nil, application.NewNotFound("emergency contact not found")
	}

	if phone != contact.Phone {
		duplicate, err := h.emergencyRepository.FindByUserIDAndPhone(ctx, cmd.UserID, phone)
		if err != nil {
			return nil, err
		}

		if duplicate != nil && duplicate.ID != contact.ID {
			return nil, application.NewConflict("emergency contact with this phone number already exists")
		}
	}

	now := h.clock().UTC()

	wasPrimary := contact.IsPrimary

	contact.Name = name
	contact.Relationship = relationship
	contact.Phone = phone
	contact.CountryCode = cmd.CountryCode
	contact.UpdatedAt = now

	// is_primary is deliberately left untouched by this write. Setting it true
	// here while another contact is still flagged would momentarily give the
	// user two primaries, which the one-primary-per-user unique index rejects.
	// SetPrimary below performs the demote and promote in one transaction.
	contact.IsPrimary = wasPrimary

	if err := h.emergencyRepository.Update(ctx, contact); err != nil {
		return nil, err
	}

	if cmd.IsPrimary && !wasPrimary {
		if err := h.emergencyRepository.SetPrimary(ctx, cmd.UserID, contact.ID, now); err != nil {
			return nil, err
		}
		contact.IsPrimary = true
	}

	// Clearing the flag on the current primary is intentionally a no-op: a user
	// must always have one, so demoting is done by promoting a different
	// contact rather than by unsetting this one.

	return EmergencyContactToDTO(contact), nil
}

// DeleteEmergencyContactHandler removes a contact, refusing to remove the last
// one.
type DeleteEmergencyContactHandler struct {
	emergencyRepository repository.EmergencyContactRepository
	clock               func() time.Time
}

func NewDeleteEmergencyContactHandler(
	emergencyRepository repository.EmergencyContactRepository,
	clock func() time.Time,
) *DeleteEmergencyContactHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &DeleteEmergencyContactHandler{
		emergencyRepository: emergencyRepository,
		clock:               clock,
	}
}

func (h *DeleteEmergencyContactHandler) Handle(
	ctx context.Context,
	cmd command.DeleteEmergencyContactCommand,
) (struct{}, error) {
	if cmd.UserID <= 0 {
		return struct{}{}, application.NewValidation("user id is required")
	}

	publicID := strings.TrimSpace(cmd.PublicID)
	if publicID == "" {
		return struct{}{}, application.NewValidation("emergency contact id is required")
	}

	contacts, err := h.emergencyRepository.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return struct{}{}, err
	}

	var target *int64
	for _, c := range contacts {
		if c.PublicID == publicID {
			id := c.ID
			target = &id
			break
		}
	}

	if target == nil {
		return struct{}{}, application.NewNotFound("emergency contact not found")
	}

	// A user with zero contacts reads as un-onboarded everywhere downstream,
	// so refuse rather than silently regressing their account state.
	if len(contacts) == 1 {
		return struct{}{}, application.NewConflict("cannot delete your only emergency contact")
	}

	if err := h.emergencyRepository.DeleteByPublicID(ctx, cmd.UserID, publicID); err != nil {
		return struct{}{}, err
	}

	// If the deleted contact was primary, promote the oldest survivor so the
	// user is never left without one.
	remaining, err := h.emergencyRepository.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return struct{}{}, err
	}

	hasPrimary := false
	for _, c := range remaining {
		if c.IsPrimary {
			hasPrimary = true
			break
		}
	}

	if !hasPrimary && len(remaining) > 0 {
		if err := h.emergencyRepository.SetPrimary(
			ctx, cmd.UserID, remaining[0].ID, h.clock().UTC(),
		); err != nil {
			return struct{}{}, err
		}
	}

	return struct{}{}, nil
}
