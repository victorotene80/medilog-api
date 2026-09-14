package sms

import (
	"context"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application/contracts"
)

// unavailableSender stands in when no SMS provider can deliver.
//
// It reports failure rather than success. The previous no-op returned nil, so
// every OTP request answered "OTP sent successfully" with nothing dispatched —
// and because OTP is the delivery channel for password reset, passwordless login
// and onboarding verification, an operator who enabled SMS without provider
// credentials had a silently unrecoverable account system that looked healthy.
//
// reason is carried into the error so the failure names its own cause instead of
// surfacing as an opaque 500.
type unavailableSender struct {
	reason string
}

// NewUnavailableSender returns an SMSSender that always fails with reason.
func NewUnavailableSender(reason string) contracts.SMSSender {
	return &unavailableSender{reason: reason}
}

func (s *unavailableSender) Send(_ context.Context, _, _ string) error {
	return fmt.Errorf("sms delivery unavailable: %s", s.reason)
}
