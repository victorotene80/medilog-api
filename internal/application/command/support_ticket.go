package command

type CreateSupportTicketCommand struct {
	UserID   int64
	Subject  string
	Category string
	Priority string
	Message  string
}

type AddSupportMessageCommand struct {
	TicketID int64
	UserID   int64
	Message  string
}

type GetSupportTicketQuery struct {
	TicketID int64
	UserID   int64
}

type ListSupportTicketsQuery struct {
	UserID int64
}
