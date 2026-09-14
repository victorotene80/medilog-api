package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type UserModel struct {
	ID                    int64      `gorm:"column:id;primaryKey;autoIncrement"`
	PublicID              string     `gorm:"column:public_id;type:uuid;default:gen_random_uuid()"`
	Email                 *string    `gorm:"column:email"`
	Phone                 *string    `gorm:"column:phone"`
	FirstName             string     `gorm:"column:first_name;not null"`
	LastName              string     `gorm:"column:last_name;not null"`
	AvatarURL             *string    `gorm:"column:avatar_url"`
	DateOfBirth           *time.Time `gorm:"column:date_of_birth"`
	Sex                   *int       `gorm:"column:sex"`
	BloodType             *string    `gorm:"column:blood_type"`
	CountryCode           *string    `gorm:"column:country_code"`
	PasswordHash          *string    `gorm:"column:password_hash"`
	Status                string     `gorm:"column:status;not null"`
	Role                  string     `gorm:"column:role;not null;default:user"`
	EmailVerifiedAt       *time.Time `gorm:"column:email_verified_at"`
	PhoneVerifiedAt       *time.Time `gorm:"column:phone_verified_at"`
	IsOnboardingCompleted bool       `gorm:"column:is_onboarding_completed;not null"`
	PasswordChangedAt     *time.Time `gorm:"column:password_changed_at"`
	FailedLoginAttempts   *int       `gorm:"column:failed_login_attempts;default:0"`
	LockedUntil           *time.Time `gorm:"column:locked_until"`
	LastLoginAt           *time.Time `gorm:"column:last_login_at"`
	LastLoginIP           *string    `gorm:"column:last_login_ip"`
	LastActiveAt          *time.Time `gorm:"column:last_active_at"`
	DeletedAt             *time.Time `gorm:"column:deleted_at"`
	CreatedAt             time.Time  `gorm:"column:created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at"`
}

func (UserModel) TableName() string {
	return "users"
}

func UserModelToEntity(m UserModel) (*entities.User, error) {
	status, err := valueobjects.NewUserStatus(m.Status)
	if err != nil {
		return nil, err
	}

	role, err := valueobjects.NewUserRole(m.Role)
	if err != nil {
		role = valueobjects.UserRoleUser
	}

	var sex *valueobjects.Sex
	if m.Sex != nil {
		value, err := valueobjects.NewSex(*m.Sex)
		if err != nil {
			return nil, err
		}
		sex = &value
	}

	var bloodType *valueobjects.BloodType
	if m.BloodType != nil {
		value, err := valueobjects.NewBloodType(*m.BloodType)
		if err != nil {
			return nil, err
		}
		bloodType = &value
	}

	var countryCode *valueobjects.CountryCode
	if m.CountryCode != nil {
		value, err := valueobjects.NewCountryCode(*m.CountryCode)
		if err != nil {
			return nil, err
		}
		countryCode = &value
	}

	failedLoginAttempts := 0
	if m.FailedLoginAttempts != nil {
		failedLoginAttempts = *m.FailedLoginAttempts
	}

	return &entities.User{
		ID:                    m.ID,
		PublicID:              m.PublicID,
		Email:                 m.Email,
		Phone:                 m.Phone,
		FirstName:             m.FirstName,
		LastName:              m.LastName,
		AvatarURL:             m.AvatarURL,
		DateOfBirth:           m.DateOfBirth,
		Sex:                   sex,
		BloodType:             bloodType,
		CountryCode:           countryCode,
		PasswordHash:          m.PasswordHash,
		Status:                status,
		Role:                  role,
		EmailVerifiedAt:       m.EmailVerifiedAt,
		PhoneVerifiedAt:       m.PhoneVerifiedAt,
		IsOnboardingCompleted: m.IsOnboardingCompleted,
		PasswordChangedAt:     m.PasswordChangedAt,
		FailedLoginAttempts:   failedLoginAttempts,
		LockedUntil:           m.LockedUntil,
		LastLoginAt:           m.LastLoginAt,
		LastLoginIP:           m.LastLoginIP,
		LastActiveAt:          m.LastActiveAt,
		DeletedAt:             m.DeletedAt,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	}, nil
}

func UserEntityToModel(user entities.User) *UserModel {
	var sex *int
	if user.Sex != nil {
		value := int(*user.Sex)
		sex = &value
	}

	var bloodType *string
	if user.BloodType != nil {
		value := user.BloodType.String()
		bloodType = &value
	}

	var countryCode *string
	if user.CountryCode != nil {
		value := user.CountryCode.String()
		countryCode = &value
	}

	status := user.Status.String()
	if status == "" {
		status = "active"
	}

	role := user.Role.String()
	if role == "" {
		role = "user"
	}

	failedLoginAttempts := user.FailedLoginAttempts

	return &UserModel{
		ID:                    user.ID,
		PublicID:              user.PublicID,
		Email:                 user.Email,
		Phone:                 user.Phone,
		FirstName:             user.FirstName,
		LastName:              user.LastName,
		AvatarURL:             user.AvatarURL,
		DateOfBirth:           user.DateOfBirth,
		Sex:                   sex,
		BloodType:             bloodType,
		CountryCode:           countryCode,
		PasswordHash:          user.PasswordHash,
		Status:                status,
		Role:                  role,
		EmailVerifiedAt:       user.EmailVerifiedAt,
		PhoneVerifiedAt:       user.PhoneVerifiedAt,
		IsOnboardingCompleted: user.IsOnboardingCompleted,
		PasswordChangedAt:     user.PasswordChangedAt,
		FailedLoginAttempts:   &failedLoginAttempts,
		LockedUntil:           user.LockedUntil,
		LastLoginAt:           user.LastLoginAt,
		LastLoginIP:           user.LastLoginIP,
		LastActiveAt:          user.LastActiveAt,
		DeletedAt:             user.DeletedAt,
		CreatedAt:             user.CreatedAt,
		UpdatedAt:             user.UpdatedAt,
	}
}
