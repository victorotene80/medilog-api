package handlers

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

const countriesCacheKey = "reference:countries:all"

type GetCountriesHandler struct {
	countryRepo repository.CountryRepository
	cache       contracts.Cache[string, []dto.CountryDTO]
}

func NewGetCountriesHandler(
	countryRepo repository.CountryRepository,
	cache contracts.Cache[string, []dto.CountryDTO],
) *GetCountriesHandler {
	return &GetCountriesHandler{
		countryRepo: countryRepo,
		cache:       cache,
	}
}

func (h *GetCountriesHandler) Handle(
	ctx context.Context,
	q query.GetCountriesQuery,
) ([]dto.CountryDTO, error) {
	cached, err := h.cache.Get(ctx, countriesCacheKey)
	if err == nil && cached != nil {
		return *cached, nil
	}

	countries, err := h.countryRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.CountryDTO, len(countries))
	for i, country := range countries {
		result[i] = dto.CountryDTO{
			Code:     country.Code,
			Name:     country.Name,
			DialCode: country.DialCode,
		}
	}

	_ = h.cache.Set(ctx, countriesCacheKey, &result)

	return result, nil
}
