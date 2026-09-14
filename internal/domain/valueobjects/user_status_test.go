package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUserStatusValid(t *testing.T) {
	tests := []struct {
		input string
		want  UserStatus
	}{
		{"active", UserStatusActive},
		{"suspended", UserStatusSuspended},
		{"deleted", UserStatusDeleted},
		{"locked", UserStatusLocked},
		{"pending_verification", UserStatusPendingVerification},
	}
	for _, tt := range tests {
		t.Run(string(tt.want), func(t *testing.T) {
			got, err := NewUserStatus(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewUserStatusInvalid(t *testing.T) {
	_, err := NewUserStatus("banned")
	assert.Error(t, err)
}

func TestUserStatusIsActive(t *testing.T) {
	assert.True(t, UserStatusActive.IsActive())
	assert.False(t, UserStatusDeleted.IsActive())
}

func TestUserStatusIsDeleted(t *testing.T) {
	assert.True(t, UserStatusDeleted.IsDeleted())
	assert.False(t, UserStatusActive.IsDeleted())
}

func TestUserStatusIsLocked(t *testing.T) {
	assert.True(t, UserStatusLocked.IsLocked())
	assert.False(t, UserStatusActive.IsLocked())
}

func TestUserStatusIsPendingVerification(t *testing.T) {
	assert.True(t, UserStatusPendingVerification.IsPendingVerification())
	assert.False(t, UserStatusActive.IsPendingVerification())
}
