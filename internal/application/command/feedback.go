package command

type SubmitFeedbackCommand struct {
	UserID      *int64
	Rating      *int
	Title       *string
	Message     string
	AppVersion  *string
	Platform    *string
	DeviceModel *string
}

type GetFeedbackQuery struct {
	FeedbackID int64
	UserID     *int64
}
