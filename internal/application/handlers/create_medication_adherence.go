package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type LogMedicationAdherenceHandler struct {
	medications domainRepo.MedicationRepository
	logs        domainRepo.MedicationAdherenceLogRepository
	tx          appContracts.TransactionManager
}

func NewLogMedicationAdherenceHandler(
	medications domainRepo.MedicationRepository,
	logs domainRepo.MedicationAdherenceLogRepository,
	tx appContracts.TransactionManager,
) *LogMedicationAdherenceHandler {
	return &LogMedicationAdherenceHandler{medications: medications, logs: logs, tx: tx}
}

func (h *LogMedicationAdherenceHandler) Handle(ctx context.Context, cmd command.LogMedicationAdherenceCommand) (struct{}, error) {
	agg, err := h.medications.FindByID(ctx, cmd.MedicationID)
	if err != nil {
		return struct{}{}, fmt.Errorf("find medication: %w", err)
	}
	if agg == nil {
		return struct{}{}, application.NewNotFound("medication not found")
	}
	if agg.Medication.UserID != cmd.UserID {
		return struct{}{}, application.NewNotFound("medication not found")
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

	// Checked: LogAdherence enforces that the log belongs to this medication and
	// this user, and discarding it silently skipped both guards.
	if err := agg.LogAdherence(log); err != nil {
		return struct{}{}, application.NewValidation(err.Error())
	}

	// The counters and the log that justifies them commit together. Separately,
	// a failed log insert left AdherenceCount and TotalDoses already incremented
	// for a dose with no row behind it — and since nothing recomputes the
	// counters from the logs, a retry drove the displayed adherence rate
	// permanently wrong. That number is shown to the patient and fed to the AI
	// as clinical fact.
	if err := h.tx.Do(ctx, func(ctx context.Context) error {
		if err := h.medications.Update(ctx, agg); err != nil {
			return fmt.Errorf("update medication: %w", err)
		}

		if err := h.logs.Save(ctx, log); err != nil {
			return fmt.Errorf("save adherence log: %w", err)
		}

		return nil
	}); err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}
