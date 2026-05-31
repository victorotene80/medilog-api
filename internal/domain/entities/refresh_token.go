package entities

import "time"

type RefreshToken struct {
	ID                  int64
	UserID              int64
	TokenHash           string
	DeviceID            *string
	DeviceName          *string
	IPAddress           *string
	DeviceFingerprint   *string
	UserAgent           *string
	ExpiresAt           time.Time
	RevokedAt           *time.Time
	ReplacedByTokenHash *string
	DateCreated         time.Time
}

func (t *RefreshToken) IsExpired(now time.Time) bool {
	return now.After(t.ExpiresAt)
}

func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

func (t *RefreshToken) IsValid(now time.Time) bool {
	return !t.IsExpired(now) && !t.IsRevoked()
}

func (t *RefreshToken) Revoke(now time.Time, replacedBy *string) {
	t.RevokedAt = &now
	t.ReplacedByTokenHash = replacedBy
}
