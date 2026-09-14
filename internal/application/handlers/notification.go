package handlers

import (
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"

	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

type ListNotificationsHandler struct {
	notificationRepo repository.NotificationRepository
	clock            func() time.Time
}

func NewListNotificationsHandler(
	notificationRepo repository.NotificationRepository,
	clock func() time.Time,
) *ListNotificationsHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &ListNotificationsHandler{
		notificationRepo: notificationRepo,
		clock:            clock,
	}
}

func (h *ListNotificationsHandler) Handle(
	ctx context.Context,
	cmd command.ListNotificationsQuery,
) ([]*dto.NotificationResponseDTO, error) {
	notifications, err := h.notificationRepo.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("find notifications: %w", err)
	}

	result := make([]*dto.NotificationResponseDTO, 0, len(notifications))
	for _, n := range notifications {
		result = append(result, notificationToDTO(n))
	}
	return result, nil
}

// notificationToDTO keeps the domain entity inside the application layer. The
// REST layer previously built its response straight from *entities.Notification,
// which made the interfaces layer depend on the domain's internal representation.
func notificationToDTO(n *entities.Notification) *dto.NotificationResponseDTO {
	if n == nil {
		return nil
	}
	return &dto.NotificationResponseDTO{
		ID:          n.PublicID,
		UserID:      strconv.FormatInt(n.UserID, 10),
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
	}
}

type GetNotificationHandler struct {
	notificationRepo repository.NotificationRepository
	clock            func() time.Time
}

func NewGetNotificationHandler(
	notificationRepo repository.NotificationRepository,
	clock func() time.Time,
) *GetNotificationHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &GetNotificationHandler{
		notificationRepo: notificationRepo,
		clock:            clock,
	}
}

func (h *GetNotificationHandler) Handle(
	ctx context.Context,
	cmd command.GetNotificationQuery,
) (*dto.NotificationResponseDTO, error) {
	// Ownership is enforced in the query, so a notification belonging to another
	// user is reported as missing rather than forbidden — a 403 would confirm
	// the id exists.
	notification, err := h.notificationRepo.FindByPublicID(ctx, cmd.UserID, cmd.PublicID)
	if err != nil {
		return nil, fmt.Errorf("find notification: %w", err)
	}

	if notification == nil {
		return nil, application.NewNotFound("notification not found")
	}

	return notificationToDTO(notification), nil
}

type MarkNotificationReadHandler struct {
	notificationRepo repository.NotificationRepository
	clock            func() time.Time
}

func NewMarkNotificationReadHandler(
	notificationRepo repository.NotificationRepository,
	clock func() time.Time,
) *MarkNotificationReadHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &MarkNotificationReadHandler{
		notificationRepo: notificationRepo,
		clock:            clock,
	}
}

func (h *MarkNotificationReadHandler) Handle(
	ctx context.Context,
	cmd command.MarkNotificationReadCommand,
) (*dto.NotificationResponseDTO, error) {
	now := h.clock()

	notification, err := h.notificationRepo.FindByPublicID(ctx, cmd.UserID, cmd.PublicID)
	if err != nil {
		return nil, fmt.Errorf("find notification: %w", err)
	}

	if notification == nil {
		return nil, application.NewNotFound("notification not found")
	}

	notification.MarkRead(now)

	if err := h.notificationRepo.Update(ctx, notification); err != nil {
		return nil, fmt.Errorf("update notification: %w", err)
	}

	return notificationToDTO(notification), nil
}

type MarkAllNotificationsReadHandler struct {
	notificationRepo repository.NotificationRepository
	clock            func() time.Time
}

func NewMarkAllNotificationsReadHandler(
	notificationRepo repository.NotificationRepository,
	clock func() time.Time,
) *MarkAllNotificationsReadHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &MarkAllNotificationsReadHandler{
		notificationRepo: notificationRepo,
		clock:            clock,
	}
}

func (h *MarkAllNotificationsReadHandler) Handle(
	ctx context.Context,
	cmd command.MarkAllNotificationsReadCommand,
) (struct{}, error) {
	now := h.clock()

	if err := h.notificationRepo.MarkAllReadByUserID(ctx, cmd.UserID, now); err != nil {
		return struct{}{}, fmt.Errorf("mark all notifications read: %w", err)
	}

	return struct{}{}, nil
}
