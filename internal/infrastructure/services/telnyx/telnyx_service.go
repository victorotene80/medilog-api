package telnyx

import "context"

type TelnyxService interface {
	SendSMS(ctx context.Context, req SendMessageRequest) (*SendMessageResult, error)
}
