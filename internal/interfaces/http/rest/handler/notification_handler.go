package handler

import (
	"net/http"

	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/httperr"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
)

type NotificationHandler struct {
	commandBus *messaging.CommandBus
	validator  contracts.Validator
}

func NewNotificationHandler(
	commandBus *messaging.CommandBus,
	validator contracts.Validator,
) *NotificationHandler {
	return &NotificationHandler{
		commandBus: commandBus,
		validator:  validator,
	}
}

// ListNotifications godoc
//
//	@Summary     List the authenticated user's notifications
//	@Description Returns the user's in-app notification inbox, newest first.
//	             Reminders are written here by the reminder scheduler; when that
//	             scheduler is disabled the list is empty.
//	@Tags        Notifications
//	@Produce     json
//	@Security    BearerAuth
//	@Success     200 {object} response.APIResponse[response.ListNotificationsResponse]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Router      /notifications [get]
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	cmd := command.ListNotificationsQuery{
		UserID: userID,
	}

	result, err := messaging.Execute[
		command.ListNotificationsQuery,
		[]*dto.NotificationResponseDTO,
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(
			w,
			httperr.StatusFrom(err),
			"LIST_NOTIFICATIONS_FAILED",
			"Could not list notifications",
			err,
		)
		return
	}

	var notifications []*response.NotificationResponse
	for _, n := range result {
		notifications = append(notifications, &response.NotificationResponse{
			ID:          n.ID,
			UserID:      n.UserID,
			Title:       n.Title,
			Body:        n.Body,
			Type:        n.Type,
			Channel:     n.Channel,
			Status:      n.Status,
			ImageURL:    n.ImageURL,
			ScheduledAt: n.ScheduledAt,
			SentAt:      n.SentAt,
			ReadAt:      n.ReadAt,
			Metadata:    n.Metadata,
			CreatedAt:   n.CreatedAt,
		})
	}

	resp := response.ListNotificationsResponse{
		Notifications: notifications,
	}

	response.Success[response.ListNotificationsResponse](w, http.StatusOK,
		"NOTIFICATIONS_FETCHED", "Notifications retrieved successfully",
		&resp,
	)
}

// GetNotification godoc
//
//	@Summary     Get a single notification
//	@Description Fetches one notification by its public UUID — the same value the
//	             list endpoint returns as `id`. A value that is not a UUID returns
//	             400; a well-formed but unknown one returns 404.
//	@Tags        Notifications
//	@Produce     json
//	@Security    BearerAuth
//	@Param       notificationId path string true "Notification public id (UUID)"
//	@Success     200 {object} response.APIResponse[response.NotificationResponse]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Failure     404 {object} response.APIResponse[response.EmptyData]
//	@Router      /notifications/{notificationId} [get]
func (h *NotificationHandler) GetNotification(w http.ResponseWriter, r *http.Request) {
	publicID, ok := publicIDParam(w, r, "notificationId", "Invalid notification ID")
	if !ok {
		return
	}

	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	cmd := command.GetNotificationQuery{
		PublicID: publicID,
		UserID:   userID,
	}

	result, err := messaging.Execute[
		command.GetNotificationQuery,
		*dto.NotificationResponseDTO,
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(
			w,
			httperr.StatusFrom(err),
			"GET_NOTIFICATION_FAILED",
			"Could not get notification",
			err,
		)
		return
	}

	if result == nil {
		response.Error(w, http.StatusNotFound, "NOTIFICATION_NOT_FOUND", "Notification not found", nil)
		return
	}

	resp := &response.NotificationResponse{
		ID:          result.ID,
		UserID:      result.UserID,
		Title:       result.Title,
		Body:        result.Body,
		Type:        result.Type,
		Channel:     result.Channel,
		Status:      result.Status,
		ImageURL:    result.ImageURL,
		ScheduledAt: result.ScheduledAt,
		SentAt:      result.SentAt,
		ReadAt:      result.ReadAt,
		Metadata:    result.Metadata,
		CreatedAt:   result.CreatedAt,
	}

	response.Success[response.NotificationResponse](w, http.StatusOK,
		"NOTIFICATION_FETCHED", "Notification retrieved successfully",
		resp,
	)
}

// MarkRead godoc
//
//	@Summary     Mark a notification as read
//	@Description Marks one notification read, addressed by its public UUID.
//	@Tags        Notifications
//	@Produce     json
//	@Security    BearerAuth
//	@Param       notificationId path string true "Notification public id (UUID)"
//	@Success     200 {object} response.APIResponse[response.NotificationResponse]
//	@Failure     400 {object} response.APIResponse[response.EmptyData]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Failure     404 {object} response.APIResponse[response.EmptyData]
//	@Router      /notifications/{notificationId}/read [patch]
func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	publicID, ok := publicIDParam(w, r, "notificationId", "Invalid notification ID")
	if !ok {
		return
	}

	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	cmd := command.MarkNotificationReadCommand{
		PublicID: publicID,
		UserID:   userID,
	}

	result, err := messaging.Execute[
		command.MarkNotificationReadCommand,
		*dto.NotificationResponseDTO,
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err),
			"MARK_READ_FAILED",
			"Could not mark notification as read",
			err,
		)
		return
	}

	if result == nil {
		response.Error(w, http.StatusNotFound, "NOTIFICATION_NOT_FOUND", "Notification not found", nil)
		return
	}

	resp := &response.NotificationResponse{
		ID:          result.ID,
		UserID:      result.UserID,
		Title:       result.Title,
		Body:        result.Body,
		Type:        result.Type,
		Channel:     result.Channel,
		Status:      result.Status,
		ImageURL:    result.ImageURL,
		ScheduledAt: result.ScheduledAt,
		SentAt:      result.SentAt,
		ReadAt:      result.ReadAt,
		Metadata:    result.Metadata,
		CreatedAt:   result.CreatedAt,
	}

	response.Success[response.NotificationResponse](w, http.StatusOK, "NOTIFICATION_UPDATED", "Notification marked as read", resp)
}

// MarkAllRead godoc
//
//	@Summary     Mark every notification as read
//	@Description Clears the unread badge by marking all of the user's unread
//	             notifications as read.
//	@Tags        Notifications
//	@Produce     json
//	@Security    BearerAuth
//	@Success     200 {object} response.APIResponse[response.EmptyData]
//	@Failure     401 {object} response.APIResponse[response.EmptyData]
//	@Router      /notifications/read-all [patch]
func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	cmd := command.MarkAllNotificationsReadCommand{
		UserID: userID,
	}

	_, err := messaging.Execute[
		command.MarkAllNotificationsReadCommand,
		struct{},
	](h.commandBus, r.Context(), cmd)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err),
			"MARK_ALL_READ_FAILED",
			"Could not mark all notifications as read",
			err,
		)
		return
	}

	response.Success[struct{}](w, http.StatusOK, "OK", "All notifications marked as read", nil)
}
