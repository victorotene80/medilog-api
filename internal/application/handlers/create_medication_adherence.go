package handlers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type LogMedicationAdherenceHandler struct {
	medications domainRepo.MedicationRepository
	logs        domainRepo.MedicationAdherenceLogRepository
}

func NewLogMedicationAdherenceHandler(
	medications domainRepo.MedicationRepository,
	logs domainRepo.MedicationAdherenceLogRepository,
) *LogMedicationAdherenceHandler {
	return &LogMedicationAdherenceHandler{medications: medications, logs: logs}
}

func (h *LogMedicationAdherenceHandler) Handle(ctx context.Context, cmd command.LogMedicationAdherenceCommand) (struct{}, error) {
	agg, err := h.medications.FindByID(ctx, cmd.MedicationID)
	if err != nil {
		return struct{}{}, fmt.Errorf("find medication: %w", err)
	}
	if agg == nil {
		return struct{}{}, errors.New("medication not found")
	}
	if agg.Medication.UserID != cmd.UserID {
		return struct{}{}, errors.New("medication not found")
	}

	status, err := valueobjects.NewAdherenceStatus(cmd.Status)
	if err != nil {
		return struct{}{}, fmt.Errorf("invalid status: %w", err)
	}

	now := time.Now().UTC()
	log := &entities.MedicationAdherenceLog{
		MedicationID: cmd.MedicationID,
		UserID:       cmd.UserID,
		ScheduledAt:  cmd.ScheduledAt,
		Status:       status,
		Note:         cmd.Note,
		CreatedAt:    now,
	}

	if status == valueobjects.AdherenceStatusTaken {
		log.MarkTaken(now)
	}

	agg.LogAdherence(log)

	if err := h.medications.Update(ctx, agg); err != nil {
		return struct{}{}, fmt.Errorf("update medication: %w", err)
	}

	if err := h.logs.Save(ctx, log); err != nil {
		return struct{}{}, fmt.Errorf("save adherence log: %w", err)
	}

	return struct{}{}, nil
}
