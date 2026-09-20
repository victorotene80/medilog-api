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
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

// stubMedicationRepo records whether the handler reached persistence. Only Save
// is exercised here; the rest satisfy the interface.
type stubMedicationRepo struct {
	saved int
}

func (s *stubMedicationRepo) FindByID(context.Context, int64) (*aggregates.MedicationAggregate, error) {
	return nil, nil
}

func (s *stubMedicationRepo) FindByPublicID(context.Context, int64, string) (*aggregates.MedicationAggregate, error) {
	return nil, nil
}

func (s *stubMedicationRepo) FindByUserID(context.Context, int64) ([]*aggregates.MedicationAggregate, error) {
	return nil, nil
}

func (s *stubMedicationRepo) FindActiveByUserID(context.Context, int64, time.Time) ([]*aggregates.MedicationAggregate, error) {
	return nil, nil
}

func (s *stubMedicationRepo) Save(context.Context, *aggregates.MedicationAggregate) error {
	s.saved++
	return nil
}

func (s *stubMedicationRepo) Update(context.Context, *aggregates.MedicationAggregate) error {
	return nil
}

func (s *stubMedicationRepo) Delete(context.Context, int64) error { return nil }

var _ repository.MedicationRepository = (*stubMedicationRepo)(nil)

// once and weekly are anchored to start_date — dayMatchesFrequency returns
// false for both when it is nil, so without this guard the medication saves and
// then never produces a reminder, with nothing logged.
func TestCreateMedication_RequiresStartDateForAnchoredFrequencies(t *testing.T) {
	tests := []struct {
		name      string
		frequency *string
		startDate *time.Time
		wantSaved bool
	}{
		{"once without start date is rejected", strPtr("once"), nil, false},
		{"weekly without start date is rejected", strPtr("weekly"), nil, false},
		{"once with start date is accepted", strPtr("once"), timePtr(time.Now()), true},
		{"weekly with start date is accepted", strPtr("weekly"), timePtr(time.Now()), true},

		// Every other frequency repeats on its own cadence and needs no anchor.
		{"daily without start date is accepted", strPtr("daily"), nil, true},
		{"twice_daily without start date is accepted", strPtr("twice_daily"), nil, true},
		{"as_needed without start date is accepted", strPtr("as_needed"), nil, true},
		{"no frequency at all is accepted", nil, nil, true},

		// NewMedicationFrequency maps an empty string to (nil, nil), so the
		// nil-guard in the check has to hold for this case.
		{"empty frequency is accepted", strPtr(""), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubMedicationRepo{}
			h := NewCreateMedicationHandler(repo)

			_, err := h.Handle(context.Background(), command.CreateMedicationCommand{
				UserID:    1,
				Name:      "Metformin",
				Frequency: tt.frequency,
				StartDate: tt.startDate,
				Times: []command.MedicationTimeInput{
					{TimeValue: "08:00"},
				},
			})

			if tt.wantSaved {
				require.NoError(t, err)
				assert.Equal(t, 1, repo.saved)
				return
			}

			require.Error(t, err)
			assert.Zero(t, repo.saved, "a rejected medication must not be persisted")

			// 422 rather than 500: the client can fix this by sending a date.
			var appErr *application.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, application.KindValidation, appErr.Kind)
			assert.Contains(t, appErr.Message, "start_date is required")
		})
	}
}
