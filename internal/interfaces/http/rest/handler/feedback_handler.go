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

type FeedbackHandler struct {
	commandBus *messaging.CommandBus
	validator  contracts.Validator
}

func NewFeedbackHandler(
	commandBus *messaging.CommandBus,
	validator contracts.Validator,
) *FeedbackHandler {
	return &FeedbackHandler{
		commandBus: commandBus,
		validator:  validator,
	}
}

func (h *FeedbackHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[request.SubmitFeedbackRequest](w, r, h.validator)
	if !ok {
		return
	}

	var userID *int64
	if uid, ok := UserIDFrom(r.Context()); ok && uid > 0 {
		userID = &uid
	}

	cmd := command.SubmitFeedbackCommand{
		UserID:      userID,
		Rating:      req.Rating,
		Title:       req.Title,
		Message:     req.Message,
		AppVersion:  req.AppVersion,
		Platform:    req.Platform,
		DeviceModel: req.DeviceModel,
	}

	_, err := messaging.Execute[
		command.SubmitFeedbackCommand,
		*dto.FeedbackResponseDTO,
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(
			w,
			httperr.StatusFrom(err),
			"SUBMIT_FEEDBACK_FAILED",
			"Could not submit feedback",
			err,
		)
		return
	}

	response.Success(
		w,
		http.StatusCreated,
		"FEEDBACK_SUBMITTED",
		"Feedback submitted successfully",
		(*struct{})(nil),
	)
}

func (h *FeedbackHandler) GetFeedback(w http.ResponseWriter, r *http.Request) {
	feedbackID, err := strconv.ParseInt(chi.URLParam(r, "feedbackId"), 10, 64)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"INVALID_ID",
			"Invalid feedback ID",
			nil,
		)
		return
	}

	var userID *int64
	if uid, ok := UserIDFrom(r.Context()); ok && uid > 0 {
		userID = &uid
	}

	cmd := command.GetFeedbackQuery{
		FeedbackID: feedbackID,
		UserID:     userID,
	}

	result, err := messaging.Execute[
		command.GetFeedbackQuery,
		*dto.FeedbackResponseDTO,
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(
			w,
			httperr.StatusFrom(err),
			"GET_FEEDBACK_FAILED",
			"Could not get feedback",
			err,
		)
		return
	}

	response.Success(
		w,
		http.StatusOK,
		"FEEDBACK_FETCHED",
		"Feedback retrieved successfully",
		result,
	)
}
