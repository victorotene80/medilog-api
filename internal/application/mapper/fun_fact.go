package mapper

import (
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

func FunFactToDTO(fact *entities.FunFact) dto.FunFactDTO {
	if fact == nil {
		return dto.FunFactDTO{}
	}

	return dto.FunFactDTO{
		ID:                fact.ID,
		Title:             fact.Title,
		Text:              fact.Text,
		Category:          fact.Category,
		TargetCountryCode: fact.TargetCountryCode,
		TargetAgeMin:      fact.TargetAgeMin,
		TargetAgeMax:      fact.TargetAgeMax,
		AllergyCategory:   fact.AllergyCategory,
		IsActive:          fact.IsActive,
		CreatedAt:         fact.CreatedAt,
	}
}
