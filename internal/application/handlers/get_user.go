package handlers

import (
	"context"
	"errors"
	"strconv"

	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

type GetUserHandler struct {
	userRepo repository.UserAggregateRepository
}

func NewGetUserHandler(userRepo repository.UserAggregateRepository) *GetUserHandler {
	return &GetUserHandler{userRepo: userRepo}
}

func (h *GetUserHandler) Handle(
	ctx context.Context,
	q query.GetUserQuery,
) (*dto.GetUserDTO, error) {
	if q.ID <= 0 {
		return nil, errors.New("user id is required")
	}

	agg, err := h.userRepo.FindByID(ctx, q.ID)
	if err != nil || agg == nil || agg.User == nil {
		return nil, err
	}

	var (
		sex       string
		bloodType string
		country   string
		height    *float64
		weight    *float64
	)

	if agg.User.Sex != nil {
		sex = agg.User.Sex.String()
	}

	if agg.User.BloodType != nil {
		bloodType = agg.User.BloodType.String()
	}

	if agg.User.CountryCode != nil {
		country = agg.User.CountryCode.String()
	}

	if agg.Profile != nil {
		height = agg.Profile.Height
		weight = agg.Profile.Weight
	}

	emergencyContact := mapPrimaryEmergencyContact(agg.EmergencyContacts)

	return &dto.GetUserDTO{
		ID:                  strconv.FormatInt(agg.User.ID, 10),
		Email:               agg.User.Email,
		Phone:               agg.User.Phone,
		FirstName:           agg.User.FirstName,
		LastName:            agg.User.LastName,
		DOB:                 agg.User.DateOfBirth,
		Sex:                 sex,
		BloodType:           bloodType,
		Height:              height,
		Weight:              weight,
		Country:             country,
		AvatarURL:           agg.User.AvatarURL,
		CountryCode:         country,
		EmergencyContact:    emergencyContact,
		Status:              agg.User.Status.String(),
		OnboardingCompleted: agg.User.IsOnboardingCompleted,
		LastLoginAt:         agg.User.LastLoginAt,
		CreatedAt:           agg.User.CreatedAt,
		UpdatedAt:           agg.User.UpdatedAt,
	}, nil
}

func mapPrimaryEmergencyContact(
	contacts []*entities.EmergencyContact,
) *dto.GetUserEmergencyContactDTO {
	if len(contacts) == 0 {
		return nil
	}

	var selected *entities.EmergencyContact
	for _, c := range contacts {
		if c != nil && c.IsPrimary {
			selected = c
			break
		}
	}

	if selected == nil {
		for _, c := range contacts {
			if c != nil {
				selected = c
				break
			}
		}
	}

	if selected == nil {
		return nil
	}

	return &dto.GetUserEmergencyContactDTO{
		ID:           strconv.FormatInt(selected.ID, 10),
		Name:         selected.Name,
		Relationship: selected.Relationship,
		Phone:        selected.Phone,
		IsPrimary:    selected.IsPrimary,
	}
}
