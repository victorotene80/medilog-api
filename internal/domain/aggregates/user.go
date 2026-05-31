package aggregates

import (
	"errors"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/events/types"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type UserAggregate struct {
	*AggregateRoot
	User              *entities.User
	Profile           *entities.UserProfile
	EmergencyContacts []*entities.EmergencyContact
	Allergies         []*entities.UserAllergy
}

func NewUserAggregate(user *entities.User) *UserAggregate {
	agg := &UserAggregate{
		AggregateRoot:     NewAggregateRoot(user.ID, 0),
		User:              user,
		Profile:           entities.NewDefaultUserProfile(user.ID),
		EmergencyContacts: make([]*entities.EmergencyContact, 0),
		Allergies:         make([]*entities.UserAllergy, 0),
	}
	agg.RaiseEvent(types.NewUserCreatedEvent(
		user.ID,
		user.Email,
		user.Phone,
		user.FirstName,
		user.LastName,
		user.Status.String(),
	))
	return agg
}

func RestoreUserAggregate(
	user *entities.User,
	profile *entities.UserProfile,
	contacts []*entities.EmergencyContact,
	allergies []*entities.UserAllergy,
	version int,
) *UserAggregate {
	return &UserAggregate{
		AggregateRoot:     NewAggregateRoot(user.ID, version),
		User:              user,
		Profile:           profile,
		EmergencyContacts: contacts,
		Allergies:         allergies,
	}
}

func (a *UserAggregate) LinkAuthProvider(
	provider string,
	providerUID string,
	email *string,
	isPrimary bool,
	now time.Time,
) *entities.UserAuthProvider {
	p := entities.NewUserAuthProvider(
		a.User.ID,
		provider,
		providerUID,
		email,
		isPrimary,
	)
	a.User.UpdatedAt = now
	a.RaiseEvent(types.NewUserAuthProviderLinkedEvent(
		a.User.ID,
		0, // ID written back after persist
		provider,
		providerUID,
		isPrimary,
	))
	return p
}

func (a *UserAggregate) VerifyEmail(now time.Time) error {
	if a.User.Email == nil {
		return errors.New("user has no email to verify")
	}
	a.User.EmailVerifiedAt = &now
	a.User.UpdatedAt = now
	a.RaiseEvent(types.NewUserEmailVerifiedEvent(a.User.ID, a.User.Email))
	return nil
}

func (a *UserAggregate) VerifyPhone(now time.Time) error {
	if a.User.Phone == nil {
		return errors.New("user has no phone to verify")
	}
	a.User.PhoneVerifiedAt = &now
	a.User.UpdatedAt = now
	a.RaiseEvent(types.NewUserPhoneVerifiedEvent(a.User.ID, a.User.Phone))
	return nil
}

func (a *UserAggregate) UpdateContact(email, phone *string, now time.Time) {
	a.User.Email = email
	a.User.Phone = phone
	a.User.UpdatedAt = now
	a.RaiseEvent(types.NewUserContactUpdatedEvent(a.User.ID, email, phone))
}

func (a *UserAggregate) ChangePassword(newHash string, now time.Time) {
	a.User.PasswordHash = &newHash
	a.User.PasswordChangedAt = &now
	a.User.UpdatedAt = now
	a.RaiseEvent(types.NewUserPasswordChangedEvent(a.User.ID))
}

func (a *UserAggregate) ChangeStatus(newStatus valueobjects.UserStatus, now time.Time) {
	old := a.User.Status.String()
	a.User.Status = newStatus
	a.User.UpdatedAt = now
	a.RaiseEvent(types.NewUserStatusChangedEvent(a.User.ID, old, newStatus.String()))
}

func (a *UserAggregate) IsLocked(now time.Time) bool {
	if a.User.LockedUntil == nil {
		return false
	}
	return now.Before(*a.User.LockedUntil)
}

func (a *UserAggregate) RecordFailedLogin(lockedUntil *time.Time, now time.Time) {
	a.User.IncrementFailedLogins()
	a.User.LockedUntil = lockedUntil
	a.User.UpdatedAt = now
	if lockedUntil != nil {
		a.ChangeStatus(valueobjects.UserStatusLocked, now)
	}
}

func (a *UserAggregate) RecordSuccessfulLogin(ip string, now time.Time) {
	a.User.ResetFailedLogins()
	a.User.LastLoginAt = &now
	a.User.LastLoginIP = &ip
	a.User.LastActiveAt = &now
	a.User.UpdatedAt = now
	a.RaiseEvent(types.NewUserLoggedInEvent(a.User.ID, a.User.Email))
}

func (a *UserAggregate) SoftDelete(now time.Time) error {
	if a.User.IsDeleted() {
		return errors.New("user is already deleted")
	}
	a.User.DeletedAt = &now
	a.User.UpdatedAt = now
	a.ChangeStatus(valueobjects.UserStatusDeleted, now)
	return nil
}

func (a *UserAggregate) CompleteOnboarding(now time.Time) error {
	if a.User.IsOnboardingCompleted {
		return errors.New("onboarding already completed")
	}
	a.User.IsOnboardingCompleted = true
	a.User.UpdatedAt = now
	a.RaiseEvent(types.NewUserOnboardingCompletedEvent(a.User.ID))
	return nil
}

func (a *UserAggregate) AttachProfile(p *entities.UserProfile) {
	a.Profile = p
}

func (a *UserAggregate) UpdateProfile(
	weight, height *float64,
	weightUnit, tempUnit string,
	now time.Time,
) {
	if a.Profile == nil {
		a.Profile = entities.NewDefaultUserProfile(a.User.ID)
	}

	a.Profile.Weight = weight
	a.Profile.Height = height

	if strings.TrimSpace(weightUnit) != "" {
		a.Profile.WeightUnit = weightUnit
	}

	if strings.TrimSpace(tempUnit) != "" {
		a.Profile.TemperatureUnit = tempUnit
	}

	a.User.UpdatedAt = now
	a.RaiseEvent(types.NewUserProfileUpdatedEvent(a.User.ID))
}

func (a *UserAggregate) AddEmergencyContact(c *entities.EmergencyContact) {
	if c.IsPrimary {
		for _, existing := range a.EmergencyContacts {
			existing.IsPrimary = false
		}
	}
	a.EmergencyContacts = append(a.EmergencyContacts, c)
}

func (a *UserAggregate) RemoveEmergencyContact(contactID int64) error {
	for i, c := range a.EmergencyContacts {
		if c.ID == contactID {
			a.EmergencyContacts = append(
				a.EmergencyContacts[:i],
				a.EmergencyContacts[i+1:]...,
			)
			return nil
		}
	}
	return errors.New("emergency contact not found")
}

func (a *UserAggregate) AddAllergy(allergy *entities.UserAllergy) {
	a.Allergies = append(a.Allergies, allergy)
}

func (a *UserAggregate) RemoveAllergy(allergyID int64) error {
	for i, al := range a.Allergies {
		if al.ID == allergyID {
			a.Allergies = append(a.Allergies[:i], a.Allergies[i+1:]...)
			return nil
		}
	}
	return errors.New("allergy not found on user")
}
