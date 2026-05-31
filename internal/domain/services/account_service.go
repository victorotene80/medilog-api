package services

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/services/policy"
)

type AccountLockService struct {
	policy policy.AccountLockPolicy
}

func NewAccountLockService(p policy.AccountLockPolicy) *AccountLockService {
	return &AccountLockService{policy: p}
}

func (s *AccountLockService) ShouldLock(failedAttempts int) bool {
	return s.policy.ShouldLockAccount(failedAttempts)
}

func (s *AccountLockService) ComputeLockedUntil(now time.Time) time.Time {
	return s.policy.ComputeLockedUntil(now)
}

func (s *AccountLockService) IsLocked(lockedUntil *time.Time, now time.Time) bool {
	if lockedUntil == nil {
		return false
	}
	return s.policy.IsStillLocked(*lockedUntil, now)
}

func (s *AccountLockService) ComputeLockExpiry(attemptCount int, now time.Time) *time.Time {
	if attemptCount < s.policy.MaxFailedAttempts {
		return nil
	}
	t := now.Add(s.policy.LockDuration)
	return &t
}
