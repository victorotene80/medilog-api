package handler

import (
	"net/http"

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

// GetQuota godoc
//
//	@Summary     Get the caller's AI question quota
//	@Description Returns how many AI questions the user has used and has left,
//	             and when the daily window next resets. The window is reset
//	             lazily on read, so this endpoint is always current.
//	@Tags        AI Conversations
//	@Produce     json
//	@Security    BearerAuth
//	@Success     200 {object} response.APIResponse[response.AIQuotaResponse]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Failure     404 {object} response.APIResponse[response.EmptyData]
//	@Router      /ai/quota [get]
func (h *AIHandler) GetQuota(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	result, err := messaging.Execute[query.GetAIQuotaQuery, *dto.AIQuotaDTO](
		h.commandBus, r.Context(), query.GetAIQuotaQuery{UserID: userID},
	)
	if err != nil {
		logAndRespond(
			w,
			httperr.StatusFrom(err),
			"AI_QUOTA_FETCH_FAILED",
			"Could not retrieve AI quota",
			err,
		)
		return
	}

	if result == nil {
		response.Error(w, http.StatusNotFound, "USER_PROFILE_NOT_FOUND", "User profile not found", nil)
		return
	}

	resp := mapper.AIQuotaDTOToResponse(*result)
	response.Success(w, http.StatusOK, "AI_QUOTA_FETCHED", "AI quota retrieved", &resp)
}

// CreateConversation godoc
//
//	@Summary     Create AI conversation
//	@Description Creates an AI conversation for the authenticated user.
//	@Tags        AI Conversations
//	@Accept      json
//	@Produce     json
//	@Security    BearerAuth
//	@Param       body body request.CreateAIConversationRequest true "AI conversation payload"
//	@Success     201 {object} response.APIResponse[response.AIConversationResponse]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Router      /ai/conversations/ [post]
func (h *AIHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
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
		logAndRespond(w, httperr.StatusFrom(err), "AI_CONVERSATION_CREATE_FAILED", "Could not create AI conversation", err)
		return
	}

	resp := mapper.AIConversationDTOToResponse(result)
	response.Success[response.AIConversationResponse](w, http.StatusCreated, "AI_CONVERSATION_CREATED", "AI conversation created successfully", &resp)
}

// ListConversations godoc
//
//	@Summary     List AI conversations
//	@Description Lists AI conversations for the authenticated user.
//	@Tags        AI Conversations
//	@Produce     json
//	@Security    BearerAuth
//	@Param       active_only query bool false "Return only active conversations"
//	@Success     200 {object} response.APIResponse[[]response.AIConversationResponse]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Router      /ai/conversations/ [get]
func (h *AIHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
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
		logAndRespond(w, httperr.StatusFrom(err), "AI_CONVERSATIONS_FETCH_FAILED", "Could not fetch AI conversations", err)
		return
	}

	resp := mapper.AIConversationDTOsToResponse(result)
	response.Success[[]response.AIConversationResponse](w, http.StatusOK, "AI_CONVERSATIONS_FETCHED", "AI conversations retrieved successfully", &resp)
}

// GetConversation godoc
//
//	@Summary     Get AI conversation
//	@Description Gets one AI conversation by public ID.
//	@Tags        AI Conversations
//	@Produce     json
//	@Security    BearerAuth
//	@Param       publicId path string true "AI conversation public ID"
//	@Success     200 {object} response.APIResponse[response.AIConversationResponse]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Failure     404 {object} response.APIResponse[response.EmptyData]
//	@Router      /ai/conversations/{publicId} [get]
func (h *AIHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	publicID, ok := publicIDParam(w, r, "publicId", "AI conversation public ID is required")
	if !ok {
		return
	}

	q := query.GetAIConversationQuery{UserID: userID, ConversationPublicID: publicID}

	result, err := messaging.Execute[query.GetAIConversationQuery, *dto.AIConversationDTO](
		h.commandBus,
		r.Context(),
		q,
	)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err), "AI_CONVERSATION_FETCH_FAILED", "Could not fetch AI conversation", err)
		return
	}
	if result == nil {
		response.Error(w, http.StatusNotFound, "AI_CONVERSATION_NOT_FOUND", "AI conversation not found", nil)
		return
	}

	resp := mapper.AIConversationDTOToResponse(*result)
	response.Success[response.AIConversationResponse](w, http.StatusOK, "AI_CONVERSATION_FETCHED", "AI conversation retrieved successfully", &resp)
}

// SendMessage godoc
//
//	@Summary     Send AI message
//	@Description Sends a user message and returns the saved message plus AI reply.
//	@Tags        AI Conversations
//	@Accept      json
//	@Produce     json
//	@Security    BearerAuth
//	@Param       publicId path string                     true "AI conversation public ID"
//	@Param       body     body request.SendAIMessageRequest true "AI message payload"
//	@Success     201 {object} response.APIResponse[response.SendAIMessageResponse]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Failure     404 {object} response.APIResponse[response.EmptyData]
//	@Router      /ai/conversations/{publicId}/messages [post]
func (h *AIHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	publicID, ok := publicIDParam(w, r, "publicId", "AI conversation public ID is required")
	if !ok {
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
		logAndRespond(w, httperr.StatusFrom(err), "AI_MESSAGE_FAILED", "Could not send AI message", err)
		return
	}

	resp := mapper.SendAIMessageResultDTOToResponse(result)
	response.Success(w, http.StatusCreated, "AI_MESSAGE_SENT", "AI message sent successfully", &resp)
}

// ArchiveConversation godoc
//
//	@Summary     Archive AI conversation
//	@Description Archives one AI conversation by public ID.
//	@Tags        AI Conversations
//	@Produce     json
//	@Security    BearerAuth
//	@Param       publicId path string true "AI conversation public ID"
//	@Success     200 {object} response.APIResponse[response.EmptyData]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Failure     404 {object} response.APIResponse[response.EmptyData]
//	@Router      /ai/conversations/{publicId}/archive [patch]
func (h *AIHandler) ArchiveConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	publicID, ok := publicIDParam(w, r, "publicId", "AI conversation public ID is required")
	if !ok {
		return
	}

	cmd := command.ArchiveAIConversationCommand{UserID: userID, ConversationPublicID: publicID}

	_, err := messaging.Execute[command.ArchiveAIConversationCommand, struct{}](
		h.commandBus,
		r.Context(),
		cmd,
	)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err), "AI_CONVERSATION_ARCHIVE_FAILED", "Could not archive AI conversation", err)
		return
	}

	writeEmptySuccess(w, http.StatusOK, "AI_CONVERSATION_ARCHIVED", "AI conversation archived successfully")
}

// UpdateConversation godoc
//
//	@Summary     Update AI conversation
//	@Description Updates an AI conversation by public ID (e.g. rename title).
//	@Tags        AI Conversations
//	@Produce     json
//	@Security    BearerAuth
//	@Param       publicId path string true "AI conversation public ID"
//	@Param       body body request.UpdateAIConversationRequest true "Update payload"
//	@Success     200 {object} response.APIResponse[response.AIConversationResponse]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Failure     404 {object} response.APIResponse[response.EmptyData]
//	@Router      /ai/conversations/{publicId} [patch]
func (h *AIHandler) UpdateConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	publicID, ok := publicIDParam(w, r, "publicId", "AI conversation public ID is required")
	if !ok {
		return
	}

	req, ok := decodeAndValidate[request.UpdateAIConversationRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.UpdateAIConversationCommand{
		UserID:               userID,
		ConversationPublicID: publicID,
		Title:                req.Title,
	}

	result, err := messaging.Execute[command.UpdateAIConversationCommand, dto.AIConversationDTO](
		h.commandBus,
		r.Context(),
		cmd,
	)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err), "AI_CONVERSATION_UPDATE_FAILED", "Could not update AI conversation", err)
		return
	}

	resp := mapper.AIConversationDTOToResponse(result)
	response.Success[response.AIConversationResponse](w, http.StatusOK, "AI_CONVERSATION_UPDATED", "AI conversation updated successfully", &resp)
}
