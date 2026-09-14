package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// The attempt allowance belongs to the code. An IP-keyed rate limit bounds
// guesses per caller, which a pool of addresses defeats against a 10^6 keyspace.
func TestOTPCode_AttemptsExhaustedInvalidatesCode(t *testing.T) {
	now := time.Now().UTC()
	otp := &OTPCode{ExpiresAt: now.Add(10 * time.Minute)}

	assert.True(t, otp.IsValid(now), "a fresh code is valid")

	for i := 0; i < MaxOTPAttempts-1; i++ {
		otp.RecordFailedAttempt()
		assert.False(t, otp.AttemptsExhausted())
		assert.True(t, otp.IsValid(now), "still valid below the allowance")
	}

	otp.RecordFailedAttempt()

	assert.True(t, otp.AttemptsExhausted())
	assert.False(t, otp.IsValid(now),
		"a code guessed MaxOTPAttempts times must stop validating even before it expires")
}

func TestOTPCode_IsValidStillCoversExpiryAndUse(t *testing.T) {
	now := time.Now().UTC()

	expired := &OTPCode{ExpiresAt: now.Add(-time.Second)}
	assert.False(t, expired.IsValid(now))

	used := &OTPCode{ExpiresAt: now.Add(time.Minute), UsedAt: &now}
	assert.False(t, used.IsValid(now))
}
