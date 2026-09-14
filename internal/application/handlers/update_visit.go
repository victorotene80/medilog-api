package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

type UpdateVisitHandler struct {
	visits domainRepo.VisitRepository
}

func NewUpdateVisitHandler(visits domainRepo.VisitRepository) *UpdateVisitHandler {
	return &UpdateVisitHandler{visits: visits}
}

func (h *UpdateVisitHandler) Handle(ctx context.Context, cmd command.UpdateVisitCommand) (struct{}, error) {
	visit, err := h.visits.FindByPublicID(ctx, cmd.UserID, cmd.PublicID)
	if err != nil {
		return struct{}{}, fmt.Errorf("find visit: %w", err)
	}
	if visit == nil {
		return struct{}{}, application.NewNotFound("visit not found")
	}

	visit.HospitalName = cmd.HospitalName
	visit.Diagnosis = cmd.Diagnosis
	visit.VisitDate = cmd.VisitDate
	visit.Outcome = cmd.Outcome
	visit.MedsCount = cmd.MedsCount
	visit.Doctor = cmd.Doctor
	visit.ChiefComplaint = cmd.ChiefComplaint
	visit.Notes = cmd.Notes
	visit.BloodPressure = cmd.BloodPressure
	visit.Temperature = cmd.Temperature
	visit.Weight = cmd.Weight
	visit.Pulse = cmd.Pulse
	visit.UpdatedAt = time.Now().UTC()

	if err := h.visits.Update(ctx, visit); err != nil {
		return struct{}{}, fmt.Errorf("update visit: %w", err)
	}

	return struct{}{}, nil
}
