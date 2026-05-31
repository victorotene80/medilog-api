package bulksms

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/shared/config"
)

var _ contracts.SMSSender = (*SenderAdapter)(nil)

type SenderAdapter struct {
	service BulkSms
	cfg     config.Config
}

func NewSenderAdapter(
	service BulkSms,
	cfg config.Config,
) *SenderAdapter {
	return &SenderAdapter{
		service: service,
		cfg:     cfg,
	}
}

func (s *SenderAdapter) Send(
	ctx context.Context,
	recipient string,
	message string,
) error {
	_, err := s.service.SendSMS(ctx, BulkSmsRequest{
		To:   recipient,
		Body: message,
		From: s.cfg.BulkSms.Sender,
	})

	return err
}
