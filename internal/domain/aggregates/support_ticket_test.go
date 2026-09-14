package aggregates

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/events"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

func TestNewSupportTicketAggregate(t *testing.T) {
	ticket := &entities.SupportTicket{
		ID:     1,
		UserID: 10,
		Status: valueobjects.TicketStatusOpen,
	}
	agg := NewSupportTicketAggregate(ticket)

	assert.Equal(t, int64(1), agg.ID())
	assert.Equal(t, ticket, agg.Ticket)
	assert.Empty(t, agg.Messages)
	assert.Empty(t, agg.Attachments)
}

func TestSupportTicketAddMessage(t *testing.T) {
	ticket := &entities.SupportTicket{
		ID:     1,
		UserID: 10,
		Status: valueobjects.TicketStatusOpen,
	}
	agg := RestoreSupportTicketAggregate(ticket, nil, nil, 0)
	now := time.Now().UTC()

	senderID := int64(10)
	msg := &entities.SupportMessage{
		ID:           100,
		TicketID:     1,
		SenderUserID: &senderID,
		Message:      "Hello, I need help",
	}
	err := agg.AddMessage(msg, now)
	assert.NoError(t, err)
	assert.Len(t, agg.Messages, 1)
	assert.Equal(t, now, *agg.Ticket.LastMessageAt)

	pulled := agg.PullEvents()
	assert.Len(t, pulled, 1)
	assert.Equal(t, events.SupportTicketMessageAddedEventName, pulled[0].EventName())
}

func TestSupportTicketAddMessageClosed(t *testing.T) {
	now := time.Now().UTC()
	ticket := &entities.SupportTicket{
		ID:       1,
		UserID:   10,
		Status:   valueobjects.TicketStatusClosed,
		ClosedAt: &now,
	}
	agg := RestoreSupportTicketAggregate(ticket, nil, nil, 0)

	msg := &entities.SupportMessage{
		ID:      100,
		TicketID: 1,
		Message: "Hello",
	}
	err := agg.AddMessage(msg, time.Now().UTC())
	assert.Error(t, err)
	assert.Equal(t, "cannot add message to a closed ticket", err.Error())
}

func TestSupportTicketAddMessageEvent(t *testing.T) {
	ticket := &entities.SupportTicket{
		ID:     1,
		UserID: 10,
		Status: valueobjects.TicketStatusOpen,
	}
	agg := RestoreSupportTicketAggregate(ticket, nil, nil, 0)

	senderID := int64(10)
	userMsg := &entities.SupportMessage{
		ID:           100,
		TicketID:     1,
		SenderUserID: &senderID,
		Message:      "User message",
	}
	err := agg.AddMessage(userMsg, time.Now().UTC())
	assert.NoError(t, err)

	events := agg.PullEvents()
	assert.Len(t, events, 1)

	sysMsg := &entities.SupportMessage{
		ID:        101,
		TicketID:  1,
		SenderUserID: nil,
		Message:   "System message",
	}
	agg2 := RestoreSupportTicketAggregate(ticket, nil, nil, 0)
	err = agg2.AddMessage(sysMsg, time.Now().UTC())
	assert.NoError(t, err)
	assert.Len(t, agg2.PullEvents(), 1)
}

func TestSupportTicketClose(t *testing.T) {
	ticket := &entities.SupportTicket{
		ID:     1,
		UserID: 10,
		Status: valueobjects.TicketStatusOpen,
	}
	agg := RestoreSupportTicketAggregate(ticket, nil, nil, 0)
	now := time.Now().UTC()

	err := agg.Close(now)
	assert.NoError(t, err)
	assert.Equal(t, valueobjects.TicketStatusClosed, agg.Ticket.Status)
	assert.Equal(t, now, *agg.Ticket.ClosedAt)

	pulled := agg.PullEvents()
	assert.Len(t, pulled, 1)
	assert.Equal(t, events.SupportTicketClosedEventName, pulled[0].EventName())
}

func TestSupportTicketCloseAlreadyClosed(t *testing.T) {
	now := time.Now().UTC()
	ticket := &entities.SupportTicket{
		ID:       1,
		UserID:   10,
		Status:   valueobjects.TicketStatusClosed,
		ClosedAt: &now,
	}
	agg := RestoreSupportTicketAggregate(ticket, nil, nil, 0)

	err := agg.Close(time.Now().UTC())
	assert.Error(t, err)
	assert.Equal(t, "ticket is already closed", err.Error())
}

func TestSupportTicketEscalate(t *testing.T) {
	ticket := &entities.SupportTicket{
		ID:       1,
		UserID:   10,
		Status:   valueobjects.TicketStatusOpen,
		Priority: valueobjects.TicketPriorityNormal,
	}
	agg := RestoreSupportTicketAggregate(ticket, nil, nil, 0)

	err := agg.Escalate(time.Now().UTC())
	assert.NoError(t, err)
	assert.Equal(t, valueobjects.TicketPriorityUrgent, agg.Ticket.Priority)
}

func TestSupportTicketEscalateClosed(t *testing.T) {
	now := time.Now().UTC()
	ticket := &entities.SupportTicket{
		ID:       1,
		UserID:   10,
		Status:   valueobjects.TicketStatusClosed,
		ClosedAt: &now,
	}
	agg := RestoreSupportTicketAggregate(ticket, nil, nil, 0)

	err := agg.Escalate(time.Now().UTC())
	assert.Error(t, err)
	assert.Equal(t, "cannot escalate a closed ticket", err.Error())
}
