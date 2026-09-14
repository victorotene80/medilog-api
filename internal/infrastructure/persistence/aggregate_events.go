package persistence

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/domain/events"
	"github.com/victorotene80/medilog-api/internal/shared/requestmeta"
)

// eventSource is the part of AggregateRoot the drain needs.
type eventSource interface {
	PullEvents() []events.DomainEvent
	ClearEvents()
}

// drainAggregateEvents writes an aggregate's pending domain events to the outbox
// on the transaction that just persisted the state change, then clears them.
//
// This lives in the repository rather than in each handler because leaving it to
// the caller demonstrably failed: RaiseEvent appended to an in-memory slice that
// only PullEvents read, no repository touched it, and only three of the call
// sites that mutate an aggregate ever published. Six events were raised and
// garbage-collected — onboarding completion, password change, the account-lock
// status change, account deletion, medication completion and adherence logging —
// while the code read as a working outbox.
//
// tx is passed explicitly rather than relying on the ambient context because the
// repositories open their own transaction and the envelope must land inside it:
// that atomicity is the entire point of the outbox.
func drainAggregateEvents(
	ctx context.Context,
	tx *gorm.DB,
	publisher appContracts.MessagePublisher,
	src eventSource,
) error {
	if publisher == nil || src == nil {
		return nil
	}

	pending := src.PullEvents()
	if len(pending) == 0 {
		return nil
	}

	if err := publisher.Publish(WithTx(ctx, tx), pending, requestMetadata(ctx)); err != nil {
		return fmt.Errorf("publish aggregate events: %w", err)
	}

	src.ClearEvents()

	return nil
}

// requestMetadata carries the caller's IP, user agent and device onto every
// event drained during that request.
//
// message_name is deliberately absent: ToEnvelope resolves the routing key per
// event from the domain event name, which is what lets one drain publish a
// mixed batch correctly. Setting it here would stamp every event in the batch
// with the same key.
func requestMetadata(ctx context.Context) map[string]string {
	meta, ok := requestmeta.FromContext(ctx)
	if !ok {
		return nil
	}

	out := make(map[string]string, 3)
	if meta.IPAddress != "" {
		out["ip_address"] = meta.IPAddress
	}
	if meta.UserAgent != "" {
		out["user_agent"] = meta.UserAgent
	}
	if meta.DeviceID != "" {
		out["device_id"] = meta.DeviceID
	}

	if len(out) == 0 {
		return nil
	}

	return out
}
