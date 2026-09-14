package entities

import "time"

// MaxOTPAttempts bounds how many times one code may be guessed.
//
// The limit belongs to the code, not to the caller: an IP-keyed rate limit
// bounds guesses per address, which a pool of addresses defeats, and the
// keyspace is only 10^6.
const MaxOTPAttempts = 5

type OTPCode struct {
	ID        int64
	UserID    int64
	Recipient string
	CodeHash  string
	Channel   string
	Purpose   string
	ExpiresAt time.Time
	UsedAt    *time.Time
	Attempts  int
	CreatedAt time.Time
}

// RecordFailedAttempt counts a wrong guess. Once the allowance is spent the code
// is no longer valid, so the attacker must request a new one — which is itself
// rate limited and invalidates the previous code.
func (o *OTPCode) RecordFailedAttempt() {
	o.Attempts++
}

// AttemptsExhausted reports whether this code has been guessed too many times.
func (o *OTPCode) AttemptsExhausted() bool {
	return o.Attempts >= MaxOTPAttempts
}

func (o *OTPCode) IsExpired(now time.Time) bool {
	return now.After(o.ExpiresAt)
}

func (o *OTPCode) IsUsed() bool {
	return o.UsedAt != nil
}

func (o *OTPCode) IsValid(now time.Time) bool {
	return !o.IsExpired(now) && !o.IsUsed() && !o.AttemptsExhausted()
}

func (o *OTPCode) MarkUsed(now time.Time) {
	o.UsedAt = &now
}
