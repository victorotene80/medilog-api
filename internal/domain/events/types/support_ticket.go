package types

import "github.com/victorotene80/medilog-api/internal/domain/events"

type SupportTicketCreatedPayload struct {
	UserID       int64  `json:"user_id"`
	TicketNumber string `json:"ticket_number"`
}

type SupportTicketMessageAddedPayload struct {
	SenderType string `json:"sender_type"`
}

type SupportTicketClosedPayload struct{}

func NewSupportTicketCreatedEvent(ticketID int64, userID int64, ticketNumber string) events.DomainEvent {
	return events.NewEvent(
		events.SupportTicketCreatedEventName,
		ticketID,
		SupportTicketCreatedPayload{UserID: userID, TicketNumber: ticketNumber},
		nil,
	)
}

func NewSupportTicketMessageAddedEvent(ticketID int64, senderType string) events.DomainEvent {
	return events.NewEvent(
		events.SupportTicketMessageAddedEventName,
		ticketID,
		SupportTicketMessageAddedPayload{SenderType: senderType},
		nil,
	)
}

func NewSupportTicketClosedEvent(ticketID int64) events.DomainEvent {
	return events.NewEvent(events.SupportTicketClosedEventName, ticketID, SupportTicketClosedPayload{}, nil)
}
