package valueobjects

import "errors"

type TicketPriority string

const (
	TicketPriorityLow    TicketPriority = "low"
	TicketPriorityNormal TicketPriority = "normal"
	TicketPriorityHigh   TicketPriority = "high"
	TicketPriorityUrgent TicketPriority = "urgent"
)

func NewTicketPriority(raw string) (TicketPriority, error) {
	p := TicketPriority(raw)
	switch p {
	case TicketPriorityLow, TicketPriorityNormal,
		TicketPriorityHigh, TicketPriorityUrgent:
		return p, nil
	}
	return "", errors.New("invalid ticket priority")
}

func (p TicketPriority) String() string { return string(p) }

type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "open"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusResolved   TicketStatus = "resolved"
	TicketStatusClosed     TicketStatus = "closed"
)

func NewTicketStatus(raw string) (TicketStatus, error) {
	s := TicketStatus(raw)
	switch s {
	case TicketStatusOpen, TicketStatusInProgress,
		TicketStatusResolved, TicketStatusClosed:
		return s, nil
	}
	return "", errors.New("invalid ticket status")
}

func (s TicketStatus) String() string { return string(s) }
