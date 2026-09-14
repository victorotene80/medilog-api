package aggregates

import (
	"errors"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/events/types"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

// SupportTicketAggregate is the consistency boundary for a support ticket,
// its messages, and file attachments.
type SupportTicketAggregate struct {
	*AggregateRoot
	Ticket      *entities.SupportTicket
	Messages    []*entities.SupportMessage
	Attachments []*entities.SupportAttachment
}

func NewSupportTicketAggregate(t *entities.SupportTicket) *SupportTicketAggregate {
	return &SupportTicketAggregate{
		AggregateRoot: NewAggregateRoot(t.ID, 0),
		Ticket:        t,
		Messages:      make([]*entities.SupportMessage, 0),
		Attachments:   make([]*entities.SupportAttachment, 0),
	}
}

func RestoreSupportTicketAggregate(
	t *entities.SupportTicket,
	messages []*entities.SupportMessage,
	attachments []*entities.SupportAttachment,
	version int,
) *SupportTicketAggregate {
	return &SupportTicketAggregate{
		AggregateRoot: NewAggregateRoot(t.ID, version),
		Ticket:        t,
		Messages:      messages,
		Attachments:   attachments,
	}
}

func (a *SupportTicketAggregate) AddMessage(msg *entities.SupportMessage, now time.Time) error {
	if !a.Ticket.IsOpen() {
		return errors.New("cannot add message to a closed ticket")
	}
	a.Messages = append(a.Messages, msg)
	a.Ticket.LastMessageAt = &now
	senderType := "user"
	if msg.IsSystemMessage() {
		senderType = "system"
	}
	a.RaiseEvent(types.NewSupportTicketMessageAddedEvent(a.Ticket.ID, senderType))
	return nil
}

func (a *SupportTicketAggregate) AttachFile(attachment *entities.SupportAttachment) {
	a.Attachments = append(a.Attachments, attachment)
}

func (a *SupportTicketAggregate) Assign(agentID int64, now time.Time) {
	a.Ticket.Assign(agentID, now)
}

func (a *SupportTicketAggregate) Close(now time.Time) error {
	if !a.Ticket.IsOpen() {
		return errors.New("ticket is already closed")
	}
	a.Ticket.Close(now)
	a.RaiseEvent(types.NewSupportTicketClosedEvent(a.Ticket.ID))
	return nil
}

func (a *SupportTicketAggregate) Escalate(now time.Time) error {
	if !a.Ticket.IsOpen() {
		return errors.New("cannot escalate a closed ticket")
	}
	a.Ticket.Priority = valueobjects.TicketPriorityUrgent
	return nil
}
