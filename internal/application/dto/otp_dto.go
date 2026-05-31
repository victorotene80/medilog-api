package dto

import "time"

type RequestOTPResultDTO struct {
	Recipient string
	Channel   string
	Purpose   string
	ExpiresAt time.Time
	Message   string
}

type VerifyOTPResultDTO struct {
	UserID    string
	Recipient string
	Channel   string
	Purpose   string
	Verified  bool
	Message   string
}