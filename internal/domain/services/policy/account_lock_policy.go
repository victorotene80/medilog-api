package policy

import "time"

type AccountLockPolicy struct {
	MaxFailedAttempts int
	LockDuration      time.Duration
}

func DefaultAccountLockPolicy() AccountLockPolicy {
	return AccountLockPolicy{
		MaxFailedAttempts: 5,
		LockDuration:      15 * time.Minute,
	}
}

func (p AccountLockPolicy) ShouldLockAccount(failedAttempts int) bool {
	return failedAttempts >= p.MaxFailedAttempts
}

func (p AccountLockPolicy) ComputeLockedUntil(now time.Time) time.Time {
	return now.Add(p.LockDuration)
}

func (p AccountLockPolicy) IsStillLocked(lockedUntil time.Time, now time.Time) bool {
	return now.Before(lockedUntil)
}
