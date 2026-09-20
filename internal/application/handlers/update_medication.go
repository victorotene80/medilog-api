package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type UpdateMedicationHandler struct {
	medications domainRepo.MedicationRepository
	times       domainRepo.MedicationTimeRepository
}

func NewUpdateMedicationHandler(
	medications domainRepo.MedicationRepository,
	times domainRepo.MedicationTimeRepository,
) *UpdateMedicationHandler {
	return &UpdateMedicationHandler{medications: medications, times: times}
}

func (h *UpdateMedicationHandler) Handle(ctx context.Context, cmd command.UpdateMedicationCommand) (struct{}, error) {
	agg, err := h.medications.FindByPublicID(ctx, cmd.UserID, cmd.PublicID)
	if err != nil {
		return struct{}{}, fmt.Errorf("find medication: %w", err)
	}
	if agg == nil {
		return struct{}{}, application.NewNotFound("medication not found")
	}

	now := time.Now().UTC()
	m := agg.Medication

	m.Name = cmd.Name
	m.DrugClass = cmd.DrugClass
	m.Dosage = cmd.Dosage
	m.WithFood = cmd.WithFood
	m.PrescribedBy = cmd.PrescribedBy
	m.Facility = cmd.Facility
	m.StartDate = cmd.StartDate
	m.EndDate = cmd.EndDate
	m.Notes = cmd.Notes
	m.UpdatedAt = now

	if cmd.Frequency != nil {
		freq, err := valueobjects.NewMedicationFrequency(cmd.Frequency)
		if err != nil {
			return struct{}{}, fmt.Errorf("invalid frequency: %w", err)
		}

		// Checked on update as well as create, and against cmd.StartDate rather
		// than m.StartDate: the assignment above is unconditional, so a PUT that
		// omits start_date actively clears it. Switching an existing medication
		// to once/weekly in the same request that drops its start date would
		// otherwise silently stop its reminders.
		if freq != nil && cmd.StartDate == nil &&
			(*freq == valueobjects.FrequencyOnce || *freq == valueobjects.FrequencyWeekly) {
			return struct{}{}, application.NewValidation(
				"start_date is required when frequency is once or weekly",
			)
		}

		m.Frequency = freq
	} else {
		m.Frequency = nil
	}

	if cmd.AddedVia != nil {
		src, err := valueobjects.NewMedicationSource(*cmd.AddedVia)
		if err != nil {
			return struct{}{}, fmt.Errorf("invalid source: %w", err)
		}
		m.AddedVia = &src
	} else {
		m.AddedVia = nil
	}

	if err := h.medications.Update(ctx, agg); err != nil {
		return struct{}{}, fmt.Errorf("update medication: %w", err)
	}

	newTimes, err := parseMedicationTimes(cmd.Times, now)
	if err != nil {
		return struct{}{}, err
	}
	for _, t := range newTimes {
		t.MedicationID = m.ID
	}
	if err := h.times.ReplaceTimes(ctx, m.ID, newTimes); err != nil {
		return struct{}{}, fmt.Errorf("replace medication times: %w", err)
	}

	return struct{}{}, nil
}
