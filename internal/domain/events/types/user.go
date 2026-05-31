package types

import "github.com/victorotene80/medilog-api/internal/domain/events"

type UserCreatedPayload struct {
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Status    string  `json:"status"`
}

func NewUserCreatedEvent(userID int64, email, phone *string, firstName, lastName, status string) events.DomainEvent {
	return events.NewEvent(
		events.UserCreatedEventName,
		userID,
		UserCreatedPayload{Email: email, Phone: phone, FirstName: firstName, LastName: lastName, Status: status},
		nil,
	)
}

type UserProfileUpdatedPayload struct{}

func NewUserProfileUpdatedEvent(userID int64) events.DomainEvent {
	return events.NewEvent(events.UserProfileUpdatedEventName, userID, UserProfileUpdatedPayload{}, nil)
}

type UserContactUpdatedPayload struct {
	Email *string `json:"email,omitempty"`
	Phone *string `json:"phone,omitempty"`
}

func NewUserContactUpdatedEvent(userID int64, email, phone *string) events.DomainEvent {
	return events.NewEvent(
		events.UserContactUpdatedEventName,
		userID,
		UserContactUpdatedPayload{Email: email, Phone: phone},
		nil,
	)
}

type UserPasswordChangedPayload struct{}

func NewUserPasswordChangedEvent(userID int64) events.DomainEvent {
	return events.NewEvent(events.UserPasswordChangedEventName, userID, UserPasswordChangedPayload{}, nil)
}

type UserStatusChangedPayload struct {
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
}

func NewUserStatusChangedEvent(userID int64, oldStatus, newStatus string) events.DomainEvent {
	return events.NewEvent(
		events.UserStatusChangedEventName,
		userID,
		UserStatusChangedPayload{OldStatus: oldStatus, NewStatus: newStatus},
		nil,
	)
}

type UserEmailVerifiedPayload struct {
	Email *string `json:"email,omitempty"`
}

func NewUserEmailVerifiedEvent(userID int64, email *string) events.DomainEvent {
	return events.NewEvent(events.UserEmailVerifiedEventName, userID, UserEmailVerifiedPayload{Email: email}, nil)
}

type UserPhoneVerifiedPayload struct {
	Phone *string `json:"phone,omitempty"`
}

func NewUserPhoneVerifiedEvent(userID int64, phone *string) events.DomainEvent {
	return events.NewEvent(events.UserPhoneVerifiedEventName, userID, UserPhoneVerifiedPayload{Phone: phone}, nil)
}

type UserAuthProviderLinkedPayload struct {
	ProviderID  int64  `json:"provider_id"`
	Provider    string `json:"provider"`
	ProviderUID string `json:"provider_uid"`
	IsPrimary   bool   `json:"is_primary"`
}

func NewUserAuthProviderLinkedEvent(userID int64, providerID int64, provider, providerUID string, isPrimary bool) events.DomainEvent {
	return events.NewEvent(
		events.UserAuthProviderLinkedEventName,
		userID,
		UserAuthProviderLinkedPayload{
			ProviderID:  providerID,
			Provider:    provider,
			ProviderUID: providerUID,
			IsPrimary:   isPrimary,
		},
		nil,
	)
}

type UserOnboardingCompletedPayload struct{}

func NewUserOnboardingCompletedEvent(userID int64) events.DomainEvent {
	return events.NewEvent(events.UserOnboardingCompletedEventName, userID, UserOnboardingCompletedPayload{}, nil)
}

type UserLoggedInPayload struct {
	Email *string `json:"email,omitempty"`
}

func NewUserLoggedInEvent(userID int64, email *string) events.DomainEvent {
	return events.NewEvent(events.UserLoggedInEventName, userID, UserLoggedInPayload{Email: email}, nil)
}
