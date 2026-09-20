package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

func userAggregateWithProfile() *aggregates.UserAggregate {
	user := &entities.User{ID: 1}
	return &aggregates.UserAggregate{
		User:    user,
		Profile: entities.NewDefaultUserProfile(user.ID),
	}
}

// applyProfileFields starts with an early return listing every profile-backed
// field. Leaving a field out of that condition makes a request carrying only
// that field return 200 and change nothing — which is the bug this whole
// change exists to fix, so it gets its own regression test.
func TestApplyProfileFields_TimezoneOnlyIsNotSkipped(t *testing.T) {
	agg := userAggregateWithProfile()
	require.Equal(t, entities.DefaultTimezone, agg.Profile.Timezone)

	err := applyProfileFields(agg, command.UpdateUserCommand{
		Timezone: strPtr("Africa/Lagos"),
	})

	require.NoError(t, err)
	assert.Equal(t, "Africa/Lagos", agg.Profile.Timezone,
		"a timezone-only update must not hit the early return")
}

func TestApplyProfileFields_TimezoneValidation(t *testing.T) {
	tests := []struct {
		name      string
		timezone  string
		wantTZ    string
		wantValid bool
	}{
		{"valid iana name", "Europe/London", "Europe/London", true},
		{"trimmed", "  Africa/Lagos  ", "Africa/Lagos", true},
		{"unresolvable zone", "Mars/Olympus", "", false},
		{"empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := userAggregateWithProfile()

			err := applyProfileFields(agg, command.UpdateUserCommand{
				Timezone: strPtr(tt.timezone),
			})

			if tt.wantValid {
				require.NoError(t, err)
				assert.Equal(t, tt.wantTZ, agg.Profile.Timezone)
				return
			}

			require.Error(t, err)

			// 422, not 500: the client sent something correctable.
			var appErr *application.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, application.KindValidation, appErr.Kind)

			assert.Equal(t, entities.DefaultTimezone, agg.Profile.Timezone,
				"a rejected timezone must not be written")
		})
	}
}

// The profile row is created lazily for users who registered before profiles
// existed. A timezone update must work for them too rather than nil-panicking.
func TestApplyProfileFields_CreatesMissingProfile(t *testing.T) {
	agg := &aggregates.UserAggregate{User: &entities.User{ID: 7}}

	err := applyProfileFields(agg, command.UpdateUserCommand{
		Timezone: strPtr("Africa/Lagos"),
	})

	require.NoError(t, err)
	require.NotNil(t, agg.Profile)
	assert.Equal(t, int64(7), agg.Profile.UserID)
	assert.Equal(t, "Africa/Lagos", agg.Profile.Timezone)
}

// No profile field set at all still short-circuits, so an update that only
// touches user-level columns does not conjure a profile row.
func TestApplyProfileFields_NoProfileFieldsIsNoop(t *testing.T) {
	agg := &aggregates.UserAggregate{User: &entities.User{ID: 3}}

	err := applyProfileFields(agg, command.UpdateUserCommand{
		FirstName: strPtr("Ada"),
	})

	require.NoError(t, err)
	assert.Nil(t, agg.Profile)
}
