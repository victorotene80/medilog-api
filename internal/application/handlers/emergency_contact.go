package handlers

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

type EmergencyContactHandler struct {
	userRepository      repository.UserRepository
	emergencyRepository repository.EmergencyContactRepository
	clock               func() time.Time
}

func NewEmergencyContactHandler(
	userRepository repository.UserRepository,
	emergencyRepository repository.EmergencyContactRepository,
	clock func() time.Time,
) *EmergencyContactHandler {
	return &EmergencyContactHandler{
		userRepository:      userRepository,
		emergencyRepository: emergencyRepository,
		clock:               clock,
	}
}

func (h *EmergencyContactHandler) Handle(
	ctx context.Context,
	cmd command.EmergencyContactCommand,
) (*dto.EmergencyContactDTO, error) {
	if cmd.UserID <= 0 {
		return nil, errors.New("user id is required")
	}

	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, errors.New("emergency contact name is required")
	}

	relationship := strings.TrimSpace(cmd.Relationship)
	if relationship == "" {
		return nil, errors.New("emergency contact relationship is required")
	}

	phone := strings.TrimSpace(cmd.Phone)
	if phone == "" {
		return nil, errors.New("emergency contact phone is required")
	}

	existingContact, err := h.emergencyRepository.FindByUserIDAndPhone(ctx, cmd.UserID, phone)
	if err != nil {
		return nil, err
	}

	if existingContact != nil {
		return nil, errors.New("emergency contact with this phone number already exists")
	}

	now := h.clock().UTC()

	emergencyContact := entities.EmergencyContact{
		UserID:       cmd.UserID,
		Name:         name,
		Relationship: relationship,
		Phone:        phone,
		IsPrimary:    true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.emergencyRepository.Save(ctx, &emergencyContact); err != nil {
		return nil, err
	}

	user, err := h.userRepository.FindByID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	user.IsOnboardingCompleted = true
	user.UpdatedAt = now

	if err := h.userRepository.Update(ctx, user); err != nil {
		return nil, err
	}

	return &dto.EmergencyContactDTO{
		ID:           strconv.FormatInt(emergencyContact.ID, 10),
		UserID:       strconv.FormatInt(emergencyContact.UserID, 10),
		Name:         emergencyContact.Name,
		Relationship: emergencyContact.Relationship,
		Phone:        emergencyContact.Phone,
		IsPrimary:    emergencyContact.IsPrimary,
		CreatedAt:    emergencyContact.CreatedAt,
		UpdatedAt:    emergencyContact.UpdatedAt,
	}, nil
}
