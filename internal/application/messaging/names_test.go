package messaging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victorotene80/medilog-api/internal/domain/events"
)

// Every event an aggregate actually raises must resolve to a published routing
// key. Without one the envelope falls back to the bare domain name, which no
// queue binds to — a topic exchange discards it silently and the relay then
// marks the outbox row sent, so the loss is invisible on both sides.
func TestWireNameFor_CoversEveryRaisedEvent(t *testing.T) {
	raised := []string{
		events.UserCreatedEventName,
		events.UserLoggedInEventName,
		events.UserPasswordChangedEventName,
		events.UserOnboardingCompletedEventName,
		events.UserStatusChangedEventName,
		events.UserProfileUpdatedEventName,
		events.UserContactUpdatedEventName,
		events.UserEmailVerifiedEventName,
		events.UserPhoneVerifiedEventName,
		events.UserAuthProviderLinkedEventName,
		events.MedicationCreatedEventName,
		events.MedicationCompletedEventName,
		events.MedicationAdherenceEventName,
		events.AIMessageAddedEventName,
		events.AIConversationArchivedEventName,
		events.SupportTicketMessageAddedEventName,
		events.SupportTicketClosedEventName,
	}

	for _, name := range raised {
		t.Run(name, func(t *testing.T) {
			wire, ok := WireNameFor(name)
			assert.True(t, ok, "%q is raised by an aggregate but has no wire name", name)
			assert.NotEqual(t, name, wire, "the wire name must not be the bare domain name")
			assert.Contains(t, wire, ".v1", "wire names are versioned: they are a published contract")
		})
	}
}

func TestWireNameFor_UnknownEventIsReported(t *testing.T) {
	_, ok := WireNameFor("something.nobodyRaises")
	assert.False(t, ok, "an unmapped event must be reported, not given an invented key")
}
