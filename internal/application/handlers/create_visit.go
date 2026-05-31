package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

type CreateVisitHandler struct {
	visits domainRepo.VisitRepository
}

func NewCreateVisitHandler(visits domainRepo.VisitRepository) *CreateVisitHandler {
	return &CreateVisitHandler{visits: visits}
}

func (h *CreateVisitHandler) Handle(ctx context.Context, cmd command.CreateVisitCommand) (struct{}, error) {
	now := time.Now().UTC()

	visit := &entities.Visit{
		UserID:         cmd.UserID,
		HospitalName:   cmd.HospitalName,
		Diagnosis:      cmd.Diagnosis,
		VisitDate:      cmd.VisitDate,
		Outcome:        cmd.Outcome,
		MedsCount:      cmd.MedsCount,
		Doctor:         cmd.Doctor,
		ChiefComplaint: cmd.ChiefComplaint,
		Notes:          cmd.Notes,
		BloodPressure:  cmd.BloodPressure,
		Temperature:    cmd.Temperature,
		Weight:         cmd.Weight,
		Pulse:          cmd.Pulse,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := h.visits.Save(ctx, visit); err != nil {
		return struct{}{}, fmt.Errorf("save visit: %w", err)
	}

	return struct{}{}, nil
}
