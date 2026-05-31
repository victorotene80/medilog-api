package dto

import "time"

type SupportTicketResponseDTO struct {
	ID            string     `json:"id"`
	TicketNumber  string     `json:"ticket_number"`
	UserID        *string    `json:"user_id,omitempty"`
	Subject       string     `json:"subject"`
	Category      string     `json:"category"`
	Status        string     `json:"status"`
	Priority      string     `json:"priority"`
	AssignedTo    *string    `json:"assigned_to,omitempty"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	ClosedAt      *time.Time `json:"closed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
 
type ListSupportTicketsResponseDTO struct {
	Tickets    []*SupportTicketResponseDTO `json:"tickets"`
	NextCursor *string                     `json:"next_cursor,omitempty"`
}
 
type SupportMessageResponseDTO struct {
	ID             string    `json:"id"`
	TicketID       string    `json:"ticket_id"`
	SenderUserID   *string   `json:"sender_user_id,omitempty"`
	SenderType     string    `json:"sender_type"`
	Message        string    `json:"message"`
	IsInternalNote bool      `json:"is_internal_note"`
	CreatedAt      time.Time `json:"created_at"`
}