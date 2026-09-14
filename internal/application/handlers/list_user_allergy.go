package handlers

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

type ListUserAllergiesHandler struct {
	userAllergyRepository repository.UserAllergyRepository
}

func NewListUserAllergiesHandler(
	userAllergyRepository repository.UserAllergyRepository,
) *ListUserAllergiesHandler {
	return &ListUserAllergiesHandler{userAllergyRepository: userAllergyRepository}
}

func (h *ListUserAllergiesHandler) Handle(
	ctx context.Context,
	q query.ListUserAllergiesQuery,
) ([]dto.UserAllergyDTO, error) {
	allergies, err := h.userAllergyRepository.FindByUserID(ctx, q.UserID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.UserAllergyDTO, 0, len(allergies))

	for _, a := range allergies {
		if a == nil {
			continue
		}

		// Apply optional category filter in-process.
		// The repository returns all; filtering here keeps the repo interface
		// simple and avoids an extra method signature.
		if q.Category != nil && a.Category.Int() != *q.Category {
			continue
		}

		result = append(result, userAllergyToDTO(a))
	}

	return result, nil
}

func userAllergyToDTO(a *entities.UserAllergy) dto.UserAllergyDTO {
	d := dto.UserAllergyDTO{
		PublicID:    a.PublicID,
		AllergyID:   a.AllergyID,
		Name:        a.Name,
		Description: a.Description,
		Category:    a.Category.Int(),
		UpdatedAt:   a.UpdatedAt,
		CategoryStr: a.Category.String(),
		IsCustom:    a.IsCustom(),
		CreatedAt:   a.CreatedAt,
	}

	if a.Severity != nil {
		sv := a.Severity.Int16()
		svStr := a.Severity.String()
		d.Severity = &sv
		d.SeverityStr = &svStr
	}

	return d
}
