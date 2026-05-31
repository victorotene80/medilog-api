package types

import "github.com/victorotene80/medilog-api/internal/domain/events"

type NotificationCreatedPayload struct {
	UserID  int64  `json:"user_id"`
	Type    string `json:"type"`
	Channel string `json:"channel,omitempty"`
}

type NotificationReadPayload struct{}
type NotificationSentPayload struct{}
type NotificationFailedPayload struct {
	Error string `json:"error"`
}

func NewNotificationCreatedEvent(notificationID int64, userID int64, notifType string, channel string) events.DomainEvent {
	return events.NewEvent(
		events.NotificationCreatedEventName,
		notificationID,
		NotificationCreatedPayload{UserID: userID, Type: notifType, Channel: channel},
		nil,
	)
}

func NewNotificationReadEvent(notificationID int64) events.DomainEvent {
	return events.NewEvent(events.NotificationReadEventName, notificationID, NotificationReadPayload{}, nil)
}

func NewNotificationSentEvent(notificationID int64) events.DomainEvent {
	return events.NewEvent(events.NotificationSentEventName, notificationID, NotificationSentPayload{}, nil)
}

func NewNotificationFailedEvent(notificationID int64, errMsg string) events.DomainEvent {
	return events.NewEvent(events.NotificationFailedEventName, notificationID, NotificationFailedPayload{Error: errMsg}, nil)
}