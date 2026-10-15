package mapper

import (
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
)

func MedicationAggregateToDTO(agg *aggregates.MedicationAggregate) dto.MedicationDTO {
	m := agg.Medication

	timeDTOs := make([]dto.MedicationTimeDTO, 0, len(agg.Times))
	for _, t := range agg.Times {
		// Stored as a UTC instant on the zero date; the driver hands it back in
		// the server's local zone, which would shift it on any non-UTC host.
		timeDTOs = append(timeDTOs, dto.MedicationTimeDTO{
			ID:        t.ID,
			TimeValue: t.TimeValue.UTC().Format("15:04"),
		})
	}

	d := dto.MedicationDTO{
		ID:                 m.ID,
		PublicID:           m.PublicID,
		UserID:             m.UserID,
		Name:               m.Name,
		DrugClass:          m.DrugClass,
		Dosage:             m.Dosage,
		WithFood:           m.WithFood,
		PrescribedBy:       m.PrescribedBy,
		Facility:           m.Facility,
		RegistrationNumber: m.RegistrationNumber,
		RegCountryCode:     m.RegCountryCode,
		IsVerified:         m.IsVerified,
		StartDate:          m.StartDate,
		EndDate:            m.EndDate,
		Notes:              m.Notes,
		AdherenceRate:      m.AdherenceRate(),
		AdherenceCount:     m.AdherenceCount,
		TotalDoses:         m.TotalDoses,
		IsCompleted:        m.IsCompleted,
		CompletedDate:      m.CompletedDate,
		Times:              timeDTOs,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}

	if m.Frequency != nil {
		s := m.Frequency.String()
		d.Frequency = &s
	}

	if m.AddedVia != nil {
		s := m.AddedVia.String()
		d.AddedVia = &s
	}

	return d
}
