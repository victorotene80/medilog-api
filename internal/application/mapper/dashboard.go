package mapper

import (
	"strconv"

	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

func DashboardToDTO(d *entities.Dashboard) *dto.DashboardDTO {
	if d == nil {
		return nil
	}

	medications := make([]dto.DashboardMedicationDTO, 0, len(d.Medications))
	for _, medication := range d.Medications {
		medications = append(medications, dto.DashboardMedicationDTO{
			ID:             medication.PublicID,
			Name:           medication.Name,
			Dosage:         medication.Dosage,
			Frequency:      medication.Frequency,
			Times:          medication.Times,
			IsVerified:     medication.IsVerified,
			IsCompleted:    medication.IsCompleted,
			CompletedDate:  medication.CompletedDate,
			AdherenceCount: medication.AdherenceCount,
			TotalDoses:     medication.TotalDoses,
		})
	}

	var funFact *dto.DashboardFunFactDTO
	if d.FunFact != nil {
		funFact = &dto.DashboardFunFactDTO{Text: d.FunFact.Text}
	}

	return &dto.DashboardDTO{
		User: dto.DashboardUserDTO{
			ID:        strconv.FormatInt(d.User.ID, 10),
			FirstName: d.User.FirstName,
			LastName:  d.User.LastName,
			AvatarURL: d.User.AvatarURL,
		},
		Medications: medications,
		HealthOverview: dto.DashboardHealthOverviewDTO{
			Visits: dto.DashboardVisitOverviewDTO{
				Total:    d.HealthOverview.Visits.Total,
				Upcoming: d.HealthOverview.Visits.Upcoming,
				Monthly:  d.HealthOverview.Visits.Monthly,
			},
			Medications: dto.DashboardMedicationOverviewDTO{
				Completed: d.HealthOverview.Medications.Completed,
				Total:     d.HealthOverview.Medications.Total,
			},
			FlaggedDrugs: dto.DashboardFlaggedDrugOverviewDTO{
				Completed: d.HealthOverview.FlaggedDrugs.Completed,
				Total:     d.HealthOverview.FlaggedDrugs.Total,
			},
		},
		FunFact: funFact,
	}
}
