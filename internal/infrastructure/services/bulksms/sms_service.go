package bulksms

import "context"

type BulkSms interface {
	SendSMS(ctx context.Context, request BulkSmsRequest) (*SendMessageResponse, error)
}
