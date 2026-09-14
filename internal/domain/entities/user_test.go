package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

func TestNewUser(t *testing.T) {
	email := "john@example.com"
	phone := "+1234567890"
	pass := "hash"
	cc := valueobjects.MustCountryCode("US")
	bt, _ := valueobjects.NewBloodType("O+")
	sex := valueobjects.SexMale
	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)

	user := NewUser(&email, &phone, "John", "Doe", cc, &pass, dob, bt, sex)

	assert.Equal(t, valueobjects.UserStatusPendingVerification, user.Status)
	assert.Equal(t, valueobjects.UserRoleUser, user.Role)
	assert.False(t, user.IsOnboardingCompleted)
}

func TestNewGoogleUser(t *testing.T) {
	now := time.Now().UTC()
	user := NewGoogleUser("jane@example.com", "Jane", "Doe", "https://pic.jpg", now)

	assert.Equal(t, valueobjects.UserStatusActive, user.Status)
	assert.NotNil(t, user.EmailVerifiedAt)
	assert.Nil(t, user.PasswordHash)
}

func TestUserFullName(t *testing.T) {
	user := &User{FirstName: "John", LastName: "Doe"}
	assert.Equal(t, "John Doe", user.FullName())
}

func TestUserIsEmailVerified(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name string
		u    *User
		want bool
	}{
		{"verified", &User{EmailVerifiedAt: &now}, true},
		{"not verified", &User{EmailVerifiedAt: nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.u.IsEmailVerified())
		})
	}
}

func TestUserIsPhoneVerified(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name string
		u    *User
		want bool
	}{
		{"verified", &User{PhoneVerifiedAt: &now}, true},
		{"not verified", &User{PhoneVerifiedAt: nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.u.IsPhoneVerified())
		})
	}
}

func TestUserIsDeleted(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name string
		u    *User
		want bool
	}{
		{"deleted", &User{DeletedAt: &now}, true},
		{"not deleted", &User{DeletedAt: nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.u.IsDeleted())
		})
	}
}

func TestUserIncrementFailedLogins(t *testing.T) {
	user := &User{FailedLoginAttempts: 0}
	user.IncrementFailedLogins()
	assert.Equal(t, 1, user.FailedLoginAttempts)
	user.IncrementFailedLogins()
	assert.Equal(t, 2, user.FailedLoginAttempts)
}

func TestUserResetFailedLogins(t *testing.T) {
	now := time.Now().UTC()
	user := &User{FailedLoginAttempts: 5, LockedUntil: &now}
	user.ResetFailedLogins()
	assert.Equal(t, 0, user.FailedLoginAttempts)
	assert.Nil(t, user.LockedUntil)
}
