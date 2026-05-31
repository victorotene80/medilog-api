package telnyx

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/application/contracts"
)

var _ contracts.SMSSender = (*SenderAdapter)(nil)

type SenderAdapter struct {
	service TelnyxService
}

func NewSenderAdapter(service TelnyxService) *SenderAdapter {
	return &SenderAdapter{
		service: service,
	}
}

func (s *SenderAdapter) Send(
	ctx context.Context,
	recipient string,
	message string,
) error {
	_, err := s.service.SendSMS(ctx, SendMessageRequest{
		To:   recipient,
		Body: message,
	})
	return err
}
