package aggregates

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victorotene80/medilog-api/internal/domain/events"
)

func TestNewAggregateRoot(t *testing.T) {
	agg := NewAggregateRoot(42, 5)
	assert.Equal(t, int64(42), agg.ID())
	assert.Equal(t, 5, agg.Version())
	assert.Empty(t, agg.PullEvents())
}

func TestRaiseEvent(t *testing.T) {
	agg := NewAggregateRoot(1, 0)
	event := events.NewEvent("test.event", 1, nil, nil)
	agg.RaiseEvent(event)
	assert.Len(t, agg.PullEvents(), 1)
}

func TestPullEventsReturnsCopy(t *testing.T) {
	agg := NewAggregateRoot(1, 0)
	agg.RaiseEvent(events.NewEvent("a", 1, nil, nil))
	agg.RaiseEvent(events.NewEvent("b", 1, nil, nil))

	pulled := agg.PullEvents()
	assert.Len(t, pulled, 2)

	pulled = pulled[:1]
	assert.Len(t, pulled, 1)
	assert.Len(t, agg.PullEvents(), 2, "original slice must not be affected")
}

func TestClearEvents(t *testing.T) {
	agg := NewAggregateRoot(1, 0)
	agg.RaiseEvent(events.NewEvent("a", 1, nil, nil))
	agg.RaiseEvent(events.NewEvent("b", 1, nil, nil))
	agg.ClearEvents()
	assert.Empty(t, agg.PullEvents())
}

func TestCommitVersion(t *testing.T) {
	agg := NewAggregateRoot(1, 0)
	assert.Equal(t, 0, agg.Version())
	agg.CommitVersion(3)
	assert.Equal(t, 3, agg.Version())
}
