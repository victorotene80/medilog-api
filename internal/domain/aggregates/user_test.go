package aggregates

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/events"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

func strPtr(s string) *string { return &s }

func TestNewUserAggregate(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Email:  strPtr("test@example.com"),
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := NewUserAggregate(user)

	assert.Equal(t, int64(1), agg.ID())
	assert.Equal(t, user, agg.User)
	assert.NotNil(t, agg.Profile)
	assert.Empty(t, agg.EmergencyContacts)
	assert.Empty(t, agg.Allergies)

	// The creation event is raised only once the user has a database identity,
	// so construction alone records nothing.
	assert.Empty(t, agg.PullEvents())

	agg.RaiseCreatedEvent()
	raised := agg.PullEvents()
	assert.Len(t, raised, 1)
	assert.Equal(t, events.UserCreatedEventName, raised[0].EventName())
	assert.Equal(t, int64(1), raised[0].AggregateID())
}

func TestVerifyEmail(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Email:  strPtr("test@example.com"),
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)
	now := time.Now().UTC()

	err := agg.VerifyEmail(now)
	assert.NoError(t, err)
	assert.NotNil(t, agg.User.EmailVerifiedAt)
	assert.Equal(t, now, *agg.User.EmailVerifiedAt)

	pulled := agg.PullEvents()
	assert.Len(t, pulled, 1)
	assert.Equal(t, events.UserEmailVerifiedEventName, pulled[0].EventName())
}

func TestVerifyEmailNoEmail(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)

	err := agg.VerifyEmail(time.Now().UTC())
	assert.Error(t, err)
	assert.Equal(t, "user has no email to verify", err.Error())
}

func TestVerifyPhone(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Phone:  strPtr("+1234567890"),
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)
	now := time.Now().UTC()

	err := agg.VerifyPhone(now)
	assert.NoError(t, err)
	assert.NotNil(t, agg.User.PhoneVerifiedAt)
	assert.Equal(t, now, *agg.User.PhoneVerifiedAt)

	pulled := agg.PullEvents()
	assert.Len(t, pulled, 1)
	assert.Equal(t, events.UserPhoneVerifiedEventName, pulled[0].EventName())
}

func TestVerifyPhoneNoPhone(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)

	err := agg.VerifyPhone(time.Now().UTC())
	assert.Error(t, err)
	assert.Equal(t, "user has no phone to verify", err.Error())
}

func TestChangePassword(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)
	now := time.Now().UTC()

	agg.ChangePassword("newhash123", now)
	assert.Equal(t, "newhash123", *agg.User.PasswordHash)
	assert.Equal(t, now, *agg.User.PasswordChangedAt)
	assert.Equal(t, now, agg.User.UpdatedAt)

	pulled := agg.PullEvents()
	assert.Len(t, pulled, 1)
	assert.Equal(t, events.UserPasswordChangedEventName, pulled[0].EventName())
}

func TestChangeStatus(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)
	now := time.Now().UTC()

	agg.ChangeStatus(valueobjects.UserStatusSuspended, now)
	assert.Equal(t, valueobjects.UserStatusSuspended, agg.User.Status)
	assert.Equal(t, now, agg.User.UpdatedAt)

	pulled := agg.PullEvents()
	assert.Len(t, pulled, 1)
	assert.Equal(t, events.UserStatusChangedEventName, pulled[0].EventName())
}

func TestIsLocked(t *testing.T) {
	future := time.Now().UTC().Add(time.Hour)
	user := &entities.User{
		ID:          1,
		LockedUntil: &future,
		Status:      valueobjects.UserStatusActive,
		Role:        valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)

	assert.True(t, agg.IsLocked(time.Now().UTC()))
}

func TestIsLockedNil(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)

	assert.False(t, agg.IsLocked(time.Now().UTC()))
}

func TestRecordFailedLogin(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)
	now := time.Now().UTC()

	agg.RecordFailedLogin(nil, now)
	assert.Equal(t, 1, agg.User.FailedLoginAttempts)
	assert.Nil(t, agg.User.LockedUntil)
}

func TestRecordFailedLoginWithLock(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)
	now := time.Now().UTC()
	lockUntil := now.Add(time.Hour)

	agg.RecordFailedLogin(&lockUntil, now)
	assert.Equal(t, 1, agg.User.FailedLoginAttempts)
	assert.NotNil(t, agg.User.LockedUntil)
	assert.Equal(t, valueobjects.UserStatusLocked, agg.User.Status)
}

func TestRecordSuccessfulLogin(t *testing.T) {
	user := &entities.User{
		ID:                  1,
		FailedLoginAttempts: 3,
		LockedUntil:         timePtr(time.Now().UTC().Add(time.Hour)),
		Status:              valueobjects.UserStatusActive,
		Role:                valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)
	now := time.Now().UTC()

	agg.RecordSuccessfulLogin("127.0.0.1", now)
	assert.Equal(t, 0, agg.User.FailedLoginAttempts)
	assert.Nil(t, agg.User.LockedUntil)
	assert.Equal(t, now, *agg.User.LastLoginAt)
	assert.Equal(t, "127.0.0.1", *agg.User.LastLoginIP)
	assert.Equal(t, now, *agg.User.LastActiveAt)

	pulled := agg.PullEvents()
	assert.Len(t, pulled, 1)
	assert.Equal(t, events.UserLoggedInEventName, pulled[0].EventName())
}

func TestSoftDelete(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)
	now := time.Now().UTC()

	err := agg.SoftDelete(now)
	assert.NoError(t, err)
	assert.NotNil(t, agg.User.DeletedAt)
	assert.Equal(t, now, *agg.User.DeletedAt)
	assert.Equal(t, valueobjects.UserStatusDeleted, agg.User.Status)
}

func TestSoftDeleteAlreadyDeleted(t *testing.T) {
	now := time.Now().UTC()
	user := &entities.User{
		ID:        1,
		DeletedAt: &now,
		Status:    valueobjects.UserStatusDeleted,
		Role:      valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)

	err := agg.SoftDelete(time.Now().UTC())
	assert.Error(t, err)
	assert.Equal(t, "user is already deleted", err.Error())
}

func TestCompleteOnboarding(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)
	now := time.Now().UTC()

	err := agg.CompleteOnboarding(now)
	assert.NoError(t, err)
	assert.True(t, agg.User.IsOnboardingCompleted)

	pulled := agg.PullEvents()
	assert.Len(t, pulled, 1)
	assert.Equal(t, events.UserOnboardingCompletedEventName, pulled[0].EventName())
}

func TestCompleteOnboardingAlreadyCompleted(t *testing.T) {
	user := &entities.User{
		ID:                    1,
		IsOnboardingCompleted: true,
		Status:                valueobjects.UserStatusActive,
		Role:                  valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)

	err := agg.CompleteOnboarding(time.Now().UTC())
	assert.Error(t, err)
	assert.Equal(t, "onboarding already completed", err.Error())
}

func TestAddEmergencyContact(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)

	contact := &entities.EmergencyContact{
		ID:           10,
		Name:         "Jane Doe",
		Relationship: "Sister",
		Phone:        "+1234567890",
		IsPrimary:    false,
	}
	agg.AddEmergencyContact(contact)
	assert.Len(t, agg.EmergencyContacts, 1)
	assert.Equal(t, "Jane Doe", agg.EmergencyContacts[0].Name)
}

func TestAddEmergencyContactPrimary(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)

	existing := &entities.EmergencyContact{
		ID:        10,
		Name:      "Existing",
		IsPrimary: true,
	}
	agg.AddEmergencyContact(existing)

	newPrimary := &entities.EmergencyContact{
		ID:        11,
		Name:      "New Primary",
		IsPrimary: true,
	}
	agg.AddEmergencyContact(newPrimary)

	assert.Len(t, agg.EmergencyContacts, 2)
	assert.False(t, agg.EmergencyContacts[0].IsPrimary, "existing contact should no longer be primary")
	assert.True(t, agg.EmergencyContacts[1].IsPrimary, "new contact should be primary")
}

func TestRemoveEmergencyContact(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, []*entities.EmergencyContact{
		{ID: 10, Name: "Jane"},
		{ID: 11, Name: "John"},
	}, nil, 0)

	err := agg.RemoveEmergencyContact(10)
	assert.NoError(t, err)
	assert.Len(t, agg.EmergencyContacts, 1)
	assert.Equal(t, int64(11), agg.EmergencyContacts[0].ID)
}

func TestRemoveEmergencyContactNotFound(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, []*entities.EmergencyContact{
		{ID: 10, Name: "Jane"},
	}, nil, 0)

	err := agg.RemoveEmergencyContact(99)
	assert.Error(t, err)
	assert.Equal(t, "emergency contact not found", err.Error())
}

func TestAddAllergy(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, nil, 0)

	allergy := &entities.UserAllergy{
		ID:   10,
		Name: "Peanuts",
	}
	agg.AddAllergy(allergy)
	assert.Len(t, agg.Allergies, 1)
	assert.Equal(t, "Peanuts", agg.Allergies[0].Name)
}

func TestRemoveAllergy(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, []*entities.UserAllergy{
		{ID: 10, Name: "Peanuts"},
		{ID: 11, Name: "Shellfish"},
	}, 0)

	err := agg.RemoveAllergy(10)
	assert.NoError(t, err)
	assert.Len(t, agg.Allergies, 1)
	assert.Equal(t, int64(11), agg.Allergies[0].ID)
}

func TestRemoveAllergyNotFound(t *testing.T) {
	user := &entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}
	agg := RestoreUserAggregate(user, nil, nil, []*entities.UserAllergy{
		{ID: 10, Name: "Peanuts"},
	}, 0)

	err := agg.RemoveAllergy(99)
	assert.Error(t, err)
	assert.Equal(t, "allergy not found on user", err.Error())
}

func timePtr(t time.Time) *time.Time { return &t }

// A lockout writes both LockedUntil and the status. Clearing only the timer
// left the account permanently "locked" for every route that checks the status,
// with no code path anywhere to restore it.
func TestRecordSuccessfulLogin_ClearsLockedStatus(t *testing.T) {
	now := time.Now().UTC()
	agg := RestoreUserAggregate(&entities.User{
		ID:     1,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	}, nil, nil, nil, 0)

	lockedUntil := now.Add(15 * time.Minute)
	agg.RecordFailedLogin(&lockedUntil, now)

	if !agg.User.Status.IsLocked() {
		t.Fatalf("expected the account to be locked, got %q", agg.User.Status)
	}

	// The lock has expired; the user supplies the right password.
	agg.RecordSuccessfulLogin("127.0.0.1", now.Add(20*time.Minute))

	if agg.User.Status.IsLocked() {
		t.Fatal("status is still locked after a successful login")
	}
	if !agg.User.Status.IsActive() {
		t.Fatalf("expected active status after a successful login, got %q", agg.User.Status)
	}
	if agg.User.LockedUntil != nil {
		t.Fatal("LockedUntil should be cleared")
	}
	if agg.User.FailedLoginAttempts != 0 {
		t.Fatalf("expected the failed attempt counter to reset, got %d", agg.User.FailedLoginAttempts)
	}
}

// Logging in must not promote an account that was never locked.
func TestRecordSuccessfulLogin_LeavesOtherStatusesAlone(t *testing.T) {
	now := time.Now().UTC()

	for _, status := range []valueobjects.UserStatus{
		valueobjects.UserStatusPendingVerification,
		valueobjects.UserStatusSuspended,
	} {
		agg := RestoreUserAggregate(&entities.User{
			ID:     1,
			Status: status,
			Role:   valueobjects.UserRoleUser,
		}, nil, nil, nil, 0)
		agg.RecordSuccessfulLogin("127.0.0.1", now)

		if agg.User.Status != status {
			t.Fatalf("status %q was changed to %q by a successful login", status, agg.User.Status)
		}
	}
}
