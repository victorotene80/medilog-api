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

// TicketCategory is the user-facing category on a support ticket.
//
// support_tickets.category_id is a plain INTEGER with no FK and there is no
// categories table, so these ids are the definition rather than a cache of one.
// They are fixed: changing a value reclassifies every existing ticket. The
// handler previously hardcoded category_id = 1 and discarded the validated
// category string entirely, so every ticket reached support unclassified.
type TicketCategory string

const (
	TicketCategoryGeneral        TicketCategory = "general"
	TicketCategoryBilling        TicketCategory = "billing"
	TicketCategoryTechnical      TicketCategory = "technical"
	TicketCategoryFeatureRequest TicketCategory = "feature_request"
	TicketCategoryBugReport      TicketCategory = "bug_report"
)

var ticketCategoryIDs = map[TicketCategory]int{
	TicketCategoryGeneral:        1,
	TicketCategoryBilling:        2,
	TicketCategoryTechnical:      3,
	TicketCategoryFeatureRequest: 4,
	TicketCategoryBugReport:      5,
}

func NewTicketCategory(raw string) (TicketCategory, error) {
	c := TicketCategory(raw)
	if _, ok := ticketCategoryIDs[c]; ok {
		return c, nil
	}
	return "", errors.New("invalid ticket category")
}

// TicketCategoryFromID reverses the mapping for reads. An unknown id falls back
// to general rather than failing a fetch on data that is already stored.
func TicketCategoryFromID(id int) TicketCategory {
	for c, cid := range ticketCategoryIDs {
		if cid == id {
			return c
		}
	}
	return TicketCategoryGeneral
}

func (c TicketCategory) Int() int       { return ticketCategoryIDs[c] }
func (c TicketCategory) String() string { return string(c) }
