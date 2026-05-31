package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/httperr"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/mapper"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/request"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
)

type AIHandler struct {
	commandBus *messaging.CommandBus
	validator  appContracts.Validator
}

func NewAIHandler(
	commandBus *messaging.CommandBus,
	validator appContracts.Validator,
) *AIHandler {
	return &AIHandler{commandBus: commandBus, validator: validator}
}

func (h *AIHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	req, ok := decodeAndValidate[request.CreateAIConversationRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.CreateAIConversationCommand{
		UserID:                    userID,
		Title:                     req.Title,
		RelatedMedicationPublicID: req.RelatedMedicationPublicID,
		RelatedVisitPublicID:      req.RelatedVisitPublicID,
	}

	result, err := messaging.Execute[command.CreateAIConversationCommand, dto.AIConversationDTO](
		h.commandBus,
		r.Context(),
		cmd,
	)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "AI_CONVERSATION_CREATE_FAILED", "Could not create AI conversation", err.Error())
		return
	}

	resp := mapper.AIConversationDTOToResponse(result)
	response.Success[response.AIConversationResponse](w, http.StatusCreated, "AI_CONVERSATION_CREATED", "AI conversation created successfully", &resp)
}

func (h *AIHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	q := query.ListAIConversationsQuery{
		UserID:     userID,
		ActiveOnly: r.URL.Query().Get("active_only") == "true",
	}

	result, err := messaging.Execute[query.ListAIConversationsQuery, []dto.AIConversationDTO](
		h.commandBus,
		r.Context(),
		q,
	)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "AI_CONVERSATIONS_FETCH_FAILED", "Could not fetch AI conversations", err.Error())
		return
	}

	resp := mapper.AIConversationDTOsToResponse(result)
	response.Success[[]response.AIConversationResponse](w, http.StatusOK, "AI_CONVERSATIONS_FETCHED", "AI conversations retrieved successfully", &resp)
}

func (h *AIHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	publicID := chi.URLParam(r, "publicId")
	if publicID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "AI conversation public ID is required", nil)
		return
	}

	q := query.GetAIConversationQuery{UserID: userID, ConversationPublicID: publicID}

	result, err := messaging.Execute[query.GetAIConversationQuery, *dto.AIConversationDTO](
		h.commandBus,
		r.Context(),
		q,
	)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "AI_CONVERSATION_FETCH_FAILED", "Could not fetch AI conversation", err.Error())
		return
	}
	if result == nil {
		response.Error(w, http.StatusNotFound, "AI_CONVERSATION_NOT_FOUND", "AI conversation not found", nil)
		return
	}

	resp := mapper.AIConversationDTOToResponse(*result)
	response.Success[response.AIConversationResponse](w, http.StatusOK, "AI_CONVERSATION_FETCHED", "AI conversation retrieved successfully", &resp)
}

func (h *AIHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	publicID := chi.URLParam(r, "publicId")
	if publicID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "AI conversation public ID is required", nil)
		return
	}

	req, ok := decodeAndValidate[request.SendAIMessageRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.SendAIMessageCommand{
		UserID:               userID,
		ConversationPublicID: publicID,
		Message:              req.Message,
		Language:             req.Language,
	}

	result, err := messaging.Execute[command.SendAIMessageCommand, *dto.SendAIMessageResultDTO](
		h.commandBus,
		r.Context(),
		cmd,
	)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "AI_MESSAGE_FAILED", "Could not send AI message", err.Error())
		return
	}

	resp := mapper.SendAIMessageResultDTOToResponse(result)
	response.Success[response.SendAIMessageResponse](w, http.StatusCreated, "AI_MESSAGE_SENT", "AI message sent successfully", &resp)
}

func (h *AIHandler) ArchiveConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	publicID := chi.URLParam(r, "publicId")
	if publicID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "AI conversation public ID is required", nil)
		return
	}

	cmd := command.ArchiveAIConversationCommand{UserID: userID, ConversationPublicID: publicID}

	_, err := messaging.Execute[command.ArchiveAIConversationCommand, struct{}](
		h.commandBus,
		r.Context(),
		cmd,
	)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "AI_CONVERSATION_ARCHIVE_FAILED", "Could not archive AI conversation", err.Error())
		return
	}

	writeEmptySuccess(w, http.StatusOK, "AI_CONVERSATION_ARCHIVED", "AI conversation archived successfully")
}
