package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

// existingMedicationRepo returns one medication from FindByPublicID and counts
// the writes that follow.
type existingMedicationRepo struct {
	stubMedicationRepo
	updated int
}

func (r *existingMedicationRepo) FindByPublicID(context.Context, int64, string) (*aggregates.MedicationAggregate, error) {
	return &aggregates.MedicationAggregate{
		Medication: &entities.Medication{ID: 42, UserID: 1, Name: "Metformin"},
	}, nil
}

func (r *existingMedicationRepo) Update(context.Context, *aggregates.MedicationAggregate) error {
	r.updated++
	return nil
}

type stubMedicationTimeRepo struct {
	replaced int
}

func (s *stubMedicationTimeRepo) FindByMedicationID(context.Context, int64) ([]*entities.MedicationTime, error) {
	return nil, nil
}

func (s *stubMedicationTimeRepo) SaveAll(context.Context, []*entities.MedicationTime) error {
	return nil
}

func (s *stubMedicationTimeRepo) DeleteByMedicationID(context.Context, int64) error { return nil }

func (s *stubMedicationTimeRepo) ReplaceTimes(context.Context, int64, []*entities.MedicationTime) error {
	s.replaced++
	return nil
}

var _ repository.MedicationTimeRepository = (*stubMedicationTimeRepo)(nil)

// Update assigns m.StartDate = cmd.StartDate unconditionally, so a PUT that
// omits start_date clears it. Switching a medication to once/weekly in that
// same request would otherwise silently end its reminders, which is why the
// guard checks cmd.StartDate rather than the already-overwritten m.StartDate.
func TestUpdateMedication_RequiresStartDateForAnchoredFrequencies(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		frequency   *string
		startDate   *time.Time
		wantWritten bool
	}{
		{"once without start date is rejected", strPtr("once"), nil, false},
		{"weekly without start date is rejected", strPtr("weekly"), nil, false},
		{"once with start date is accepted", strPtr("once"), &now, true},
		{"daily without start date is accepted", strPtr("daily"), nil, true},
		{"no frequency at all is accepted", nil, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			medRepo := &existingMedicationRepo{}
			timeRepo := &stubMedicationTimeRepo{}
			h := NewUpdateMedicationHandler(medRepo, timeRepo)

			_, err := h.Handle(context.Background(), command.UpdateMedicationCommand{
				UserID:    1,
				PublicID:  "med-public-id",
				Name:      "Metformin",
				Frequency: tt.frequency,
				StartDate: tt.startDate,
				Times: []command.MedicationTimeInput{
					{TimeValue: "08:00"},
				},
			})

			if tt.wantWritten {
				require.NoError(t, err)
				assert.Equal(t, 1, medRepo.updated)
				return
			}

			require.Error(t, err)
			assert.Zero(t, medRepo.updated, "a rejected update must not be persisted")
			assert.Zero(t, timeRepo.replaced,
				"a rejected update must not reach ReplaceTimes, which would wipe the dose times")

			var appErr *application.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, application.KindValidation, appErr.Kind)
		})
	}
}
