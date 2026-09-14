package contracts

import (
	"context"
	"time"

	domainToken "github.com/victorotene80/medilog-api/internal/domain/contracts"
)

type SessionResult struct {
	SessionID    string
	AccessToken  domainToken.Token
	RefreshToken domainToken.Token
	ExpiresAt    time.Time
}

type CachedToken struct {
	UserID       string
	SessionID    string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	TokenVersion int16
}

type SessionService interface {
	Create(
		ctx context.Context,
		userID string,
		ipAddress string,
		userAgent string,
		deviceID string,
		deviceFingerprint string,
		deviceName string,
	) (SessionResult, error)

	Refresh(
		ctx context.Context,
		oldRefreshToken string,
		ipAddress string,
		userAgent string,
		deviceID string,
		deviceFingerprint string,
		deviceName string,
	) (SessionResult, error)
}
