package twilio

import "context"

type TwilioService interface {
	SendSMS(ctx context.Context, req SendMessageRequest) (*SendMessageResult, error)
	SendWhatsApp(ctx context.Context, req SendMessageRequest) (*SendMessageResult, error)
}