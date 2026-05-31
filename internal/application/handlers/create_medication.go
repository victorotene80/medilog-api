package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type CreateMedicationHandler struct {
	medications domainRepo.MedicationRepository
}

func NewCreateMedicationHandler(medications domainRepo.MedicationRepository) *CreateMedicationHandler {
	return &CreateMedicationHandler{medications: medications}
}

func (h *CreateMedicationHandler) Handle(ctx context.Context, cmd command.CreateMedicationCommand) (struct{}, error) {
	now := time.Now().UTC()

	med := &entities.Medication{
		UserID:             cmd.UserID,
		Name:               cmd.Name,
		DrugClass:          cmd.DrugClass,
		Dosage:             cmd.Dosage,
		WithFood:           cmd.WithFood,
		PrescribedBy:       cmd.PrescribedBy,
		Facility:           cmd.Facility,
		RegistrationNumber: cmd.RegistrationNumber,
		RegCountryCode:     cmd.RegCountryCode,
		StartDate:          cmd.StartDate,
		EndDate:            cmd.EndDate,
		Notes:              cmd.Notes,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if cmd.Frequency != nil {
		freq, err := valueobjects.NewMedicationFrequency(cmd.Frequency)
		if err != nil {
			return struct{}{}, fmt.Errorf("invalid frequency: %w", err)
		}
		med.Frequency = freq
	}

	if cmd.AddedVia != nil {
		src, err := valueobjects.NewMedicationSource(*cmd.AddedVia)
		if err != nil {
			return struct{}{}, fmt.Errorf("invalid source: %w", err)
		}
		med.AddedVia = &src
	}

	times, err := parseMedicationTimes(cmd.Times, now)
	if err != nil {
		return struct{}{}, err
	}

	agg := aggregates.NewMedicationAggregate(med, times)

	if err := h.medications.Save(ctx, agg); err != nil {
		return struct{}{}, fmt.Errorf("save medication: %w", err)
	}

	return struct{}{}, nil
}

func parseMedicationTimes(inputs []command.MedicationTimeInput, now time.Time) ([]*entities.MedicationTime, error) {
	times := make([]*entities.MedicationTime, 0, len(inputs))
	for _, input := range inputs {
		t, err := time.Parse("15:04", input.TimeValue)
		if err != nil {
			return nil, fmt.Errorf("invalid time value %q: expected HH:MM", input.TimeValue)
		}
		times = append(times, &entities.MedicationTime{
			TimeValue: t,
			CreatedAt: now,
		})
	}
	return times, nil
}
