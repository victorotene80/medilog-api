package handlers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/command"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
	"gorm.io/gorm"
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
		return struct{}{}, errors.New("medication not found")
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

	// Replace times: delete old, insert new.
	if err := h.times.DeleteByMedicationID(ctx, m.ID); err != nil {
		return struct{}{}, fmt.Errorf("clear medication times: %w", err)
	}

	if len(cmd.Times) > 0 {
		newTimes, err := parseMedicationTimes(cmd.Times, now)
		if err != nil {
			return struct{}{}, err
		}
		for _, t := range newTimes {
			t.MedicationID = m.ID
		}
		if err := h.times.SaveAll(ctx, newTimes); err != nil {
			return struct{}{}, fmt.Errorf("save medication times: %w", err)
		}
	}

	return struct{}{}, nil
}

// FindByPublicID is added to the repository interface — see section 7.
var _ = gorm.ErrRecordNotFound // keep gorm import tidy if unused elsewhere
