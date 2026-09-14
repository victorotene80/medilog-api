package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/httperr"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/request"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
)

type SupportTicketHandler struct {
	commandBus *messaging.CommandBus
	validator  contracts.Validator
}

func NewSupportTicketHandler(
	commandBus *messaging.CommandBus,
	validator contracts.Validator,
) *SupportTicketHandler {
	return &SupportTicketHandler{
		commandBus: commandBus,
		validator:  validator,
	}
}

func (h *SupportTicketHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[request.CreateSupportTicketRequest](w, r, h.validator)
	if !ok {
		return
	}

	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	cmd := command.CreateSupportTicketCommand{
		UserID:   userID,
		Subject:  req.Subject,
		Category: req.Category,
		Priority: req.Priority,
		Message:  req.Message,
	}

	result, err := messaging.Execute[
		command.CreateSupportTicketCommand,
		*dto.SupportTicketResponseDTO,
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err),
			"CREATE_TICKET_FAILED",
			"Could not create support ticket",
			err,
		)
		return
	}

	response.Success(w, http.StatusCreated,
		"CREATED", "Support ticket created successfully",
		result,
	)
}

func (h *SupportTicketHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	ticketID, err := strconv.ParseInt(chi.URLParam(r, "ticketId"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest,
			"BAD_REQUEST", "Invalid ticket ID",
			nil,
		)
		return
	}

	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	cmd := command.GetSupportTicketQuery{
		TicketID: ticketID,
		UserID:   userID,
	}

	result, err := messaging.Execute[
		command.GetSupportTicketQuery,
		*dto.SupportTicketResponseDTO,
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err),
			"GET_TICKET_FAILED",
			"Could not get support ticket",
			err,
		)
		return
	}

	response.Success(w, http.StatusOK,
		"OK", "Support ticket retrieved successfully",
		result,
	)
}

func (h *SupportTicketHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	cmd := command.ListSupportTicketsQuery{
		UserID: userID,
	}

	result, err := messaging.Execute[
		command.ListSupportTicketsQuery,
		*dto.ListSupportTicketsResponseDTO,
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err),
			"LIST_TICKETS_FAILED",
			"Could not list support tickets",
			err,
		)
		return
	}

	response.Success(w, http.StatusOK,
		"OK", "Support tickets retrieved successfully",
		result,
	)
}

func (h *SupportTicketHandler) AddMessage(w http.ResponseWriter, r *http.Request) {
	ticketID, err := strconv.ParseInt(chi.URLParam(r, "ticketId"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest,
			"BAD_REQUEST", "Invalid ticket ID",
			nil,
		)
		return
	}

	req, ok := decodeAndValidate[request.AddSupportMessageRequest](w, r, h.validator)
	if !ok {
		return
	}

	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	cmd := command.AddSupportMessageCommand{
		TicketID: ticketID,
		UserID:   userID,
		Message:  req.Message,
	}

	result, err := messaging.Execute[
		command.AddSupportMessageCommand,
		*dto.SupportMessageResponseDTO,
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err),
			"ADD_MESSAGE_FAILED",
			"Could not add message to support ticket",
			err,
		)
		return
	}

	response.Success(w, http.StatusCreated,
		"CREATED", "Message added successfully",
		result,
	)
}
