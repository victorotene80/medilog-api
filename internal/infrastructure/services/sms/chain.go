package sms

import (
	"context"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application/contracts"
)

type ChainSender struct {
	providers []contracts.SMSSender
}

func NewChainSender(providers ...contracts.SMSSender) contracts.SMSSender {
	return &ChainSender{providers: providers}
}

func (c *ChainSender) Send(
	ctx context.Context,
	recipient string,
	message string,
) error {
	var lastErr error

	for _, p := range c.providers {
		if err := p.Send(ctx, recipient, message); err != nil {
			lastErr = err
			continue
		}
		return nil
	}

	return fmt.Errorf("sms: all providers failed, last error: %w", lastErr)
}