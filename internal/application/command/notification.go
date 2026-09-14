package command

type ListNotificationsQuery struct {
	UserID int64
}

// PublicID is the notification's UUID, matching what the list endpoint returns
// as `id`. The internal row id is never accepted from a client.
type GetNotificationQuery struct {
	PublicID string
	UserID   int64
}

type MarkNotificationReadCommand struct {
	PublicID string
	UserID   int64
}

type MarkAllNotificationsReadCommand struct {
	UserID int64
}
