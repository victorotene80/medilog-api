package handlers

import (
	"context"
	"errors"

	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type GetAllergiesHandler struct {
	allergyRepository repository.AllergyRepository
}

func NewGetAllergiesHandler(
	allergyRepository repository.AllergyRepository,
) *GetAllergiesHandler {
	return &GetAllergiesHandler{
		allergyRepository: allergyRepository,
	}
}

func (h *GetAllergiesHandler) Handle(
	ctx context.Context,
	q query.GetAllergiesQuery,
) ([]dto.AllergyDTO, error) {
	if h.allergyRepository == nil {
		return nil, errors.New("allergy repository is required")
	}

	var allergies []*entities.Allergy
	var err error

	if q.Category != nil {
		// Declared with = rather than := on purpose: a fresh `category, err :=`
		// here shadows the outer err, so the repository error below is written to
		// the inner variable and the check after this block reads nil — a
		// database failure then returns 200 with an empty catalogue.
		category, categoryErr := valueobjects.NewAllergyCategory(*q.Category)
		if categoryErr != nil {
			return nil, categoryErr
		}

		allergies, err = h.allergyRepository.FindByCategory(ctx, category.Int())
	} else {
		allergies, err = h.allergyRepository.FindAll(ctx)
	}

	if err != nil {
		return nil, err
	}

	result := make([]dto.AllergyDTO, 0, len(allergies))

	for _, allergy := range allergies {
		if allergy == nil {
			continue
		}

		result = append(result, allergyToDTO(allergy))
	}

	return result, nil
}

func allergyToDTO(allergy *entities.Allergy) dto.AllergyDTO {
	return dto.AllergyDTO{
		ID:          allergy.ID,
		Name:        allergy.Name,
		Category:    allergy.Category.Int(),
		Description: allergy.Description,
	}
}
