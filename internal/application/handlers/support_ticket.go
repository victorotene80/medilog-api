package handlers

import (
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"

	"context"
	"fmt"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type CreateSupportTicketHandler struct {
	ticketRepo repository.SupportTicketRepository
	clock      func() time.Time
}

func NewCreateSupportTicketHandler(
	ticketRepo repository.SupportTicketRepository,
	clock func() time.Time,
) *CreateSupportTicketHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &CreateSupportTicketHandler{
		ticketRepo: ticketRepo,
		clock:      clock,
	}
}

func (h *CreateSupportTicketHandler) Handle(
	ctx context.Context,
	cmd command.CreateSupportTicketCommand,
) (*dto.SupportTicketResponseDTO, error) {
	now := h.clock()

	priority, err := valueobjects.NewTicketPriority(cmd.Priority)
	if err != nil {
		return nil, application.NewValidation(err.Error())
	}

	category, err := valueobjects.NewTicketCategory(cmd.Category)
	if err != nil {
		return nil, application.NewValidation(err.Error())
	}

	subject := strings.TrimSpace(cmd.Subject)
	if subject == "" {
		return nil, application.NewValidation("subject is required")
	}

	// Both were validated at the edge and then thrown away: category_id was
	// hardcoded to 1 and the subject was never stored, so every ticket reached
	// support with no title and in one bucket.
	ticket := &entities.SupportTicket{
		UserID:      cmd.UserID,
		CategoryID:  category.Int(),
		Status:      valueobjects.TicketStatusOpen,
		Priority:    priority,
		Description: &subject,
		CreatedAt:   now,
	}

	msg := &entities.SupportMessage{
		SenderUserID: &cmd.UserID,
		Message:      cmd.Message,
		CreatedAt:    now,
	}

	agg := aggregates.NewSupportTicketAggregate(ticket)
	if err := agg.AddMessage(msg, now); err != nil {
		return nil, application.NewValidation(err.Error())
	}

	if err := h.ticketRepo.Save(ctx, agg); err != nil {
		return nil, fmt.Errorf("save ticket: %w", err)
	}

	return supportTicketToDTO(agg), nil
}

// supportTicketToDTO is the single place a ticket becomes a response.
//
// It exists because three handlers each built this struct by hand and drifted:
// the list path omitted Subject and Category, so the support list was a column
// of blank titles while the detail view rendered correctly.
func supportTicketToDTO(agg *aggregates.SupportTicketAggregate) *dto.SupportTicketResponseDTO {
	if agg == nil || agg.Ticket == nil {
		return nil
	}

	t := agg.Ticket

	return &dto.SupportTicketResponseDTO{
		ID:            fmt.Sprintf("%d", t.ID),
		UserID:        strPtr(fmt.Sprintf("%d", t.UserID)),
		Subject:       derefString(t.Description),
		Category:      valueobjects.TicketCategoryFromID(t.CategoryID).String(),
		Status:        t.Status.String(),
		Priority:      t.Priority.String(),
		AssignedTo:    assignedToStr(t.AssignedTo),
		LastMessageAt: t.LastMessageAt,
		ClosedAt:      t.ClosedAt,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.CreatedAt,
	}
}

type GetSupportTicketHandler struct {
	ticketRepo repository.SupportTicketRepository
	clock      func() time.Time
}

func NewGetSupportTicketHandler(
	ticketRepo repository.SupportTicketRepository,
	clock func() time.Time,
) *GetSupportTicketHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &GetSupportTicketHandler{
		ticketRepo: ticketRepo,
		clock:      clock,
	}
}

func (h *GetSupportTicketHandler) Handle(
	ctx context.Context,
	cmd command.GetSupportTicketQuery,
) (*dto.SupportTicketResponseDTO, error) {
	agg, err := h.ticketRepo.FindByID(ctx, cmd.TicketID)
	if err != nil {
		return nil, fmt.Errorf("find ticket: %w", err)
	}

	if agg == nil {
		return nil, application.NewNotFound("support ticket not found")
	}

	if agg.Ticket.UserID != cmd.UserID {
		// Not-found rather than forbidden: a 403 confirms the ticket exists, and
		// ids are sequential, so it leaks the platform's ticket volume and which
		// conversations are real. Same reasoning as notification.go:67.
		return nil, application.NewNotFound("support ticket not found")
	}

	return supportTicketToDTO(agg), nil
}

type ListSupportTicketsHandler struct {
	ticketRepo repository.SupportTicketRepository
	clock      func() time.Time
}

func NewListSupportTicketsHandler(
	ticketRepo repository.SupportTicketRepository,
	clock func() time.Time,
) *ListSupportTicketsHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &ListSupportTicketsHandler{
		ticketRepo: ticketRepo,
		clock:      clock,
	}
}

func (h *ListSupportTicketsHandler) Handle(
	ctx context.Context,
	cmd command.ListSupportTicketsQuery,
) (*dto.ListSupportTicketsResponseDTO, error) {
	tickets, err := h.ticketRepo.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("find tickets: %w", err)
	}

	result := make([]*dto.SupportTicketResponseDTO, len(tickets))
	for i, agg := range tickets {
		result[i] = supportTicketToDTO(agg)
	}

	return &dto.ListSupportTicketsResponseDTO{
		Tickets: result,
	}, nil
}

type AddSupportMessageHandler struct {
	ticketRepo repository.SupportTicketRepository
	clock      func() time.Time
}

func NewAddSupportMessageHandler(
	ticketRepo repository.SupportTicketRepository,
	clock func() time.Time,
) *AddSupportMessageHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &AddSupportMessageHandler{
		ticketRepo: ticketRepo,
		clock:      clock,
	}
}

func (h *AddSupportMessageHandler) Handle(
	ctx context.Context,
	cmd command.AddSupportMessageCommand,
) (*dto.SupportMessageResponseDTO, error) {
	now := h.clock()

	agg, err := h.ticketRepo.FindByID(ctx, cmd.TicketID)
	if err != nil {
		return nil, fmt.Errorf("find ticket: %w", err)
	}

	if agg == nil {
		return nil, application.NewNotFound("support ticket not found")
	}

	if agg.Ticket.UserID != cmd.UserID {
		// Not-found rather than forbidden: a 403 confirms the ticket exists, and
		// ids are sequential, so it leaks the platform's ticket volume and which
		// conversations are real. Same reasoning as notification.go:67.
		return nil, application.NewNotFound("support ticket not found")
	}

	msg := &entities.SupportMessage{
		TicketID:     cmd.TicketID,
		SenderUserID: &cmd.UserID,
		Message:      cmd.Message,
		CreatedAt:    now,
	}

	if err := agg.AddMessage(msg, now); err != nil {
		return nil, application.NewValidation(err.Error())
	}

	if err := h.ticketRepo.Update(ctx, agg); err != nil {
		return nil, fmt.Errorf("update ticket: %w", err)
	}

	return &dto.SupportMessageResponseDTO{
		ID:           fmt.Sprintf("%d", msg.ID),
		TicketID:     fmt.Sprintf("%d", msg.TicketID),
		SenderUserID: strPtr(fmt.Sprintf("%d", *msg.SenderUserID)),
		Message:      msg.Message,
		CreatedAt:    msg.CreatedAt,
	}, nil
}

func strPtr(s string) *string {
	return &s
}

func assignedToStr(assignedTo *int64) *string {
	if assignedTo == nil {
		return nil
	}
	s := fmt.Sprintf("%d", *assignedTo)
	return &s
}
