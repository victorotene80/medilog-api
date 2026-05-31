package command

type ChangePasswordCommand struct {
	UserID      int64
	OldPassword string
	NewPassword string
}
