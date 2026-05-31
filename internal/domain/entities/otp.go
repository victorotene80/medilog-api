package entities

import "time"

type OTPCode struct {
	ID        int64
	UserID    *int64
	Recipient string
	CodeHash  string
	Channel   string
	Purpose   string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

func (o *OTPCode) IsExpired(now time.Time) bool {
	return now.After(o.ExpiresAt)
}

func (o *OTPCode) IsUsed() bool {
	return o.UsedAt != nil
}

func (o *OTPCode) IsValid(now time.Time) bool {
	return !o.IsExpired(now) && !o.IsUsed()
}

func (o *OTPCode) MarkUsed(now time.Time) {
	o.UsedAt = &now
}
