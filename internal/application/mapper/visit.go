package mapper

import (
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

func VisitToDTO(v *entities.Visit) dto.VisitDTO {
	return dto.VisitDTO{
		ID:             v.ID,
		PublicID:       v.PublicID,
		UserID:         v.UserID,
		HospitalName:   v.HospitalName,
		Diagnosis:      v.Diagnosis,
		VisitDate:      v.VisitDate,
		Outcome:        v.Outcome,
		MedsCount:      v.MedsCount,
		Doctor:         v.Doctor,
		ChiefComplaint: v.ChiefComplaint,
		Notes:          v.Notes,
		BloodPressure:  v.BloodPressure,
		Temperature:    v.Temperature,
		Weight:         v.Weight,
		Pulse:          v.Pulse,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
}

func VisitsToDTO(visits []*entities.Visit) []dto.VisitDTO {
	result := make([]dto.VisitDTO, 0, len(visits))
	for _, visit := range visits {
		result = append(result, VisitToDTO(visit))
	}
	return result
}
