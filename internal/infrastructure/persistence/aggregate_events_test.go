package persistence

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appmsg "github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/events"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
	"github.com/victorotene80/medilog-api/internal/shared/requestmeta"
)

type capturingPublisher struct {
	batches [][]events.DomainEvent
	metas   []map[string]string
	err     error
}

func (c *capturingPublisher) Publish(
	_ context.Context, e []events.DomainEvent, meta map[string]string,
) error {
	if c.err != nil {
		return c.err
	}
	c.batches = append(c.batches, e)
	c.metas = append(c.metas, meta)
	return nil
}

func (c *capturingPublisher) PublishEnvelope(context.Context, appmsg.Envelope) error { return nil }

func userAggregateWithEvent(t *testing.T) *aggregates.UserAggregate {
	t.Helper()

	agg := aggregates.NewUserAggregate(&entities.User{
		ID:     7,
		Status: valueobjects.UserStatusActive,
		Role:   valueobjects.UserRoleUser,
	})
	agg.RaiseCreatedEvent()
	require.Len(t, agg.PullEvents(), 1)

	return agg
}

// The drain is what makes the outbox structural. Before it, RaiseEvent appended
// to a slice only PullEvents read, no repository touched it, and most mutations
// raised events that were garbage-collected.
func TestDrainAggregateEvents_PublishesAndClears(t *testing.T) {
	agg := userAggregateWithEvent(t)
	pub := &capturingPublisher{}

	require.NoError(t, drainAggregateEvents(context.Background(), nil, pub, agg))

	require.Len(t, pub.batches, 1)
	assert.Len(t, pub.batches[0], 1)
	assert.Empty(t, agg.PullEvents(), "drained events must not be published twice")
}

// A failed publish must fail the write, not be swallowed — that atomicity is
// the reason the drain lives inside the repository transaction.
func TestDrainAggregateEvents_PublishFailureIsReturnedAndEventsKept(t *testing.T) {
	agg := userAggregateWithEvent(t)
	sentinel := errors.New("outbox unavailable")
	pub := &capturingPublisher{err: sentinel}

	err := drainAggregateEvents(context.Background(), nil, pub, agg)

	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	assert.Len(t, agg.PullEvents(), 1, "events must survive a failed publish")
}

func TestDrainAggregateEvents_NoEventsIsANoOp(t *testing.T) {
	agg := aggregates.NewUserAggregate(&entities.User{ID: 7, Status: valueobjects.UserStatusActive})
	pub := &capturingPublisher{}

	require.NoError(t, drainAggregateEvents(context.Background(), nil, pub, agg))
	assert.Empty(t, pub.batches, "an aggregate with nothing pending must not publish")
}

// Drained events carry the caller's request metadata, so moving publication out
// of the handlers does not lose the IP/UA/device attribution they attached.
func TestDrainAggregateEvents_CarriesRequestMetadata(t *testing.T) {
	agg := userAggregateWithEvent(t)
	pub := &capturingPublisher{}

	ctx := requestmeta.WithMeta(context.Background(), requestmeta.Meta{
		IPAddress: "203.0.113.9",
		UserAgent: "medilog-ios/1.2",
		DeviceID:  "device-42",
	})

	require.NoError(t, drainAggregateEvents(ctx, nil, pub, agg))

	require.Len(t, pub.metas, 1)
	assert.Equal(t, "203.0.113.9", pub.metas[0]["ip_address"])
	assert.Equal(t, "medilog-ios/1.2", pub.metas[0]["user_agent"])
	assert.Equal(t, "device-42", pub.metas[0]["device_id"])
	assert.NotContains(t, pub.metas[0], "message_name",
		"the routing key is resolved per event by ToEnvelope, not stamped on the batch")
}
