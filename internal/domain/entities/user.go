package entities

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type User struct {
	ID                    int64
	PublicID              string
	Email                 *string
	Phone                 *string
	FirstName             string
	LastName              string
	AvatarURL             *string
	DateOfBirth           *time.Time
	Sex                   *valueobjects.Sex
	BloodType             *valueobjects.BloodType
	CountryCode           *valueobjects.CountryCode
	PasswordHash          *string
	Status                valueobjects.UserStatus
	EmailVerifiedAt       *time.Time
	PhoneVerifiedAt       *time.Time
	IsOnboardingCompleted bool
	PasswordChangedAt     *time.Time
	FailedLoginAttempts   int
	LockedUntil           *time.Time
	LastLoginAt           *time.Time
	LastLoginIP           *string
	LastActiveAt          *time.Time
	DeletedAt             *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func NewUser(
	email *string,
	phone *string,
	firstName string,
	lastName string,
	countryCode valueobjects.CountryCode,
	password *string,
	dateOfBirth time.Time,
	bloodType valueobjects.BloodType,
	sex valueobjects.Sex,
) *User {
	now := time.Now().UTC()

	return &User{
		Email:                 email,
		Phone:                 phone,
		FirstName:             firstName,
		LastName:              lastName,
		CountryCode:           &countryCode,
		PasswordHash:          password,
		DateOfBirth:           &dateOfBirth,
		BloodType:             &bloodType,
		Sex:                   &sex,
		Status:                valueobjects.UserStatusPendingVerification,
		IsOnboardingCompleted: false,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

func NewGoogleUser(
	email string,
	firstName string,
	lastName string,
	pictureURL string,
	now time.Time,
) *User {
	return &User{
		Email:                 &email,
		Phone:                 nil,
		FirstName:             firstName,
		LastName:              lastName,
		AvatarURL:             nullableString(pictureURL),
		PasswordHash:          nil,
		DateOfBirth:           nil,
		Sex:                   nil,
		BloodType:             nil,
		CountryCode:           nil,
		Status:                valueobjects.UserStatusActive,
		EmailVerifiedAt:       &now,
		IsOnboardingCompleted: false,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

func (u *User) IsPhoneVerified() bool {
	return u.PhoneVerifiedAt != nil
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

func (u *User) IncrementFailedLogins() {
	u.FailedLoginAttempts++
}

func (u *User) ResetFailedLogins() {
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
}
