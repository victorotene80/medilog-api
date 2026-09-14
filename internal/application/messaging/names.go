package messaging

import "github.com/victorotene80/medilog-api/internal/domain/events"

// Wire names for integration events.
//
// These are the routing keys the relay publishes with and that consumers bind
// to, so they are a contract with other services and with the broker topology —
// not an internal detail. They are deliberately versioned and deliberately
// distinct from the domain event names in internal/domain/events: a domain
// event name ("user.created") describes something that happened inside this
// service, while a wire name ("auth.user.created.v1") is a published interface
// that other deployments depend on and that can only change by adding a version.
//
// Keep every publisher and every Subscribe call on these constants. They were
// previously three independently maintained lists of string literals, and they
// had already drifted: RegisterHandler published with nil metadata, so the
// envelope fell back to the bare domain name "user.created" and matched no
// binding — a topic exchange silently discards an unroutable message, and the
// relay then marked the outbox row sent.
const (
	EventUserCreated       = "auth.user.created.v1"
	EventUserLoginRecorded = "auth.user.login-recorded.v1"
	EventUserGoogleCreated = "auth.user.google-created.v1"
	EventUserGoogleLinked  = "auth.user.google-linked.v1"
	EventUserLocked        = "auth.user.locked.v1"
	EventSessionCreated    = "auth.session.created.v1"
	EventSessionRevoked    = "auth.session.revoked.v1"
	EventPasswordChanged   = "auth.password.changed.v1"
)

// Wire names for task queues.
const (
	TaskSendWelcomeEmail  = "auth.send-welcome-email.v1"
	TaskSendVerification  = "auth.send-verification.v1"
	TaskSyncAnalyticsUser = "auth.sync-analytics-user.v1"
)

// Wire names for the remaining aggregates' events. Added when the repositories
// began draining aggregates automatically: before that these events were raised
// and garbage-collected, so they had never needed a routing key.
const (
	EventUserOnboardingCompleted = "auth.user.onboarding-completed.v1"
	EventUserStatusChanged       = "auth.user.status-changed.v1"
	EventUserProfileUpdated      = "auth.user.profile-updated.v1"
	EventUserContactUpdated      = "auth.user.contact-updated.v1"
	EventUserEmailVerified       = "auth.user.email-verified.v1"
	EventUserPhoneVerified       = "auth.user.phone-verified.v1"
	EventUserAuthProviderLinked  = "auth.user.auth-provider-linked.v1"

	EventMedicationCreated         = "medication.created.v1"
	EventMedicationCompleted       = "medication.completed.v1"
	EventMedicationAdherenceLogged = "medication.adherence-logged.v1"

	EventAIMessageAdded         = "ai.conversation.message-added.v1"
	EventAIConversationArchived = "ai.conversation.archived.v1"

	EventSupportTicketMessageAdded = "support.ticket.message-added.v1"
	EventSupportTicketClosed       = "support.ticket.closed.v1"
)

// wireNames maps a domain event name to its published routing key.
//
// The two are deliberately different things: a domain event name describes what
// happened inside this service and may be renamed freely; a wire name is a
// contract other deployments bind to. Without this map an envelope falls back to
// the bare domain name, which matches no binding — a topic exchange discards an
// unroutable message silently and the relay then marks the outbox row sent.
var wireNames = map[string]string{
	events.UserCreatedEventName:             EventUserCreated,
	events.UserLoggedInEventName:            EventUserLoginRecorded,
	events.UserPasswordChangedEventName:     EventPasswordChanged,
	events.UserOnboardingCompletedEventName: EventUserOnboardingCompleted,
	events.UserStatusChangedEventName:       EventUserStatusChanged,
	events.UserProfileUpdatedEventName:      EventUserProfileUpdated,
	events.UserContactUpdatedEventName:      EventUserContactUpdated,
	events.UserEmailVerifiedEventName:       EventUserEmailVerified,
	events.UserPhoneVerifiedEventName:       EventUserPhoneVerified,
	events.UserAuthProviderLinkedEventName:  EventUserAuthProviderLinked,

	events.MedicationCreatedEventName:   EventMedicationCreated,
	events.MedicationCompletedEventName: EventMedicationCompleted,
	events.MedicationAdherenceEventName: EventMedicationAdherenceLogged,

	events.AIMessageAddedEventName:         EventAIMessageAdded,
	events.AIConversationArchivedEventName: EventAIConversationArchived,

	events.SupportTicketMessageAddedEventName: EventSupportTicketMessageAdded,
	events.SupportTicketClosedEventName:       EventSupportTicketClosed,
}

// WireNameFor returns the routing key for a domain event name. The second
// result is false when the event has no published contract yet — callers should
// treat that as "not for external consumption" rather than inventing a key.
func WireNameFor(domainEventName string) (string, bool) {
	name, ok := wireNames[domainEventName]
	return name, ok
}
