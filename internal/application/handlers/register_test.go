package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/test/testutil"
)

type inlineTx struct{}

func (inlineTx) Do(ctx context.Context, fn func(ctx context.Context) error) error { return fn(ctx) }

type acceptAllPasswords struct{}

func (acceptAllPasswords) Validate(string) error { return nil }

func TestRegister_Timezone(t *testing.T) {
	tests := []struct {
		name     string
		timezone *string
		want     string
	}{
		{"device timezone is stored", strPtr("Africa/Lagos"), "Africa/Lagos"},
		{"omitted keeps the default", nil, entities.DefaultTimezone},
		// Failing sign-up at its last step over a timezone would be worse than
		// a reminder offset the user can fix from their profile.
		{"unrecognized keeps the default", strPtr("Mars/Olympus"), entities.DefaultTimezone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(testutil.MockUserAggregateRepo)
			repo.On("ExistsByEmail", mock.Anything, "ada@example.com").Return(false, nil)
			var saved *aggregates.UserAggregate
			repo.On("Save", mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) { saved = args.Get(1).(*aggregates.UserAggregate) }).
				Return(nil)

			hasher := new(testutil.MockPasswordHasher)
			hasher.On("Hash", mock.Anything).Return("hash", nil)

			h := NewRegisterHandler(repo, inlineTx{}, hasher, nil,
				func() time.Time { return time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC) },
				acceptAllPasswords{})

			_, err := h.Handle(context.Background(), command.RegisterCommand{
				FirstName:   "Ada",
				LastName:    "Obi",
				Email:       strPtr("ada@example.com"),
				Password:    "Str0ng!pass",
				DOB:         "1990-05-01",
				Sex:         1,
				BloodType:   "O+",
				CountryCode: "NG",
				Timezone:    tt.timezone,
			})

			require.NoError(t, err)
			require.NotNil(t, saved)
			assert.Equal(t, tt.want, saved.Profile.Timezone)
		})
	}
}
