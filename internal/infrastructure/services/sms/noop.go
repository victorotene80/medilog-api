package sms

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/application/contracts"
)

type noopSender struct{}

func NewNoopSender() contracts.SMSSender { return &noopSender{} }

func (s *noopSender) Send(_ context.Context, _, _ string) error {
	return nil
}