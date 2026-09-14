package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	domainContracts "github.com/victorotene80/medilog-api/internal/domain/contracts"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/services"
	"github.com/victorotene80/medilog-api/internal/domain/services/policy"
)

var _ appContracts.SessionService = (*SessionService)(nil)

type SessionService struct {
	refreshTokenRepo repository.RefreshTokenRepository
	tokenGen         domainContracts.TokenGenerator
	hasher           *services.SessionKeyHasher
	policy           policy.SessionPolicy
	tokenCache       appContracts.Cache[string, appContracts.CachedToken]
	version          appContracts.SessionCache
	clock            func() time.Time
}

func NewSessionService(
	refreshTokenRepo repository.RefreshTokenRepository,
	tokenGen domainContracts.TokenGenerator,
	hasher *services.SessionKeyHasher,
	policy policy.SessionPolicy,
	tokenCache appContracts.Cache[string, appContracts.CachedToken],
	version appContracts.SessionCache,
	clock func() time.Time,
) *SessionService {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}
	return &SessionService{
		refreshTokenRepo: refreshTokenRepo,
		tokenGen:         tokenGen,
		hasher:           hasher,
		policy:           policy,
		tokenCache:       tokenCache,
		version:          version,
		clock:            clock,
	}
}

func (s *SessionService) Create(
	ctx context.Context,
	userID string,
	ipAddress string,
	userAgent string,
	deviceID string,
	deviceFingerprint string,
	deviceName string,
) (appContracts.SessionResult, error) {
	now := s.clock()
	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return appContracts.SessionResult{}, fmt.Errorf("invalid userID: %w", err)
	}

	activeSessions, err := s.refreshTokenRepo.FindActiveByUserID(ctx, userIDInt)
	if err != nil {
		return appContracts.SessionResult{}, fmt.Errorf("find active sessions: %w", err)
	}
	if len(activeSessions) >= s.policy.MaxConcurrentSessions {
		oldest := activeSessions[0]
		for _, sess := range activeSessions {
			if sess.DateCreated.Before(oldest.DateCreated) {
				oldest = sess
			}
		}
		oldest.Revoke(now, nil)
		if err := s.refreshTokenRepo.Update(ctx, oldest); err != nil {
			return appContracts.SessionResult{}, fmt.Errorf("revoke oldest session: %w", err)
		}

		// Revoking the row only stops that device's next refresh. Its access
		// token is validated from the session cache, which never reads the row,
		// so without this the "signed out" device keeps full access for the
		// remaining life of its token.
		if s.tokenCache != nil {
			_ = s.tokenCache.Delete(ctx, strconv.FormatInt(oldest.ID, 10))
		}
	}

	rawToken, err := services.GenerateRandomString(32)
	if err != nil {
		return appContracts.SessionResult{}, fmt.Errorf("generate session token: %w", err)
	}
	tokenHash := s.hasher.Hash(rawToken)

	refreshEntity := &entities.RefreshToken{
		UserID:      userIDInt,
		TokenHash:   tokenHash,
		ExpiresAt:   now.Add(s.policy.RefreshTokenDuration),
		DateCreated: now,
	}

	if deviceID != "" {
		refreshEntity.DeviceID = &deviceID
	}
	if deviceName != "" {
		refreshEntity.DeviceName = &deviceName
	}
	if ipAddress != "" {
		refreshEntity.IPAddress = &ipAddress
	}
	if userAgent != "" {
		refreshEntity.UserAgent = &userAgent
	}
	if deviceFingerprint != "" {
		refreshEntity.DeviceFingerprint = &deviceFingerprint
	}

	if err := s.refreshTokenRepo.Save(ctx, refreshEntity); err != nil {
		return appContracts.SessionResult{}, fmt.Errorf("save refresh token: %w", err)
	}

	sessionID := fmt.Sprintf("%d", refreshEntity.ID)

	accessToken, err := s.tokenGen.GenerateAccess(userID, sessionID, s.policy.MaxDuration)
	if err != nil {
		return appContracts.SessionResult{}, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.tokenGen.GenerateRefresh(userID, sessionID, s.policy.RefreshTokenDuration)
	if err != nil {
		return appContracts.SessionResult{}, fmt.Errorf("generate refresh token jwt: %w", err)
	}

	if s.tokenCache != nil {
		// Stamp the session version this entry was written under. Without it the
		// entry reads as version 0 forever and a later revocation cannot
		// invalidate it. Skip the write entirely if the version is unreadable —
		// an unstamped entry is worse than a cache miss.
		version, versionErr := s.currentVersion(ctx, userID)
		if versionErr == nil {
			cached := &appContracts.CachedToken{
				UserID:       userID,
				SessionID:    sessionID,
				AccessToken:  accessToken.Value,
				RefreshToken: refreshToken.Value,
				ExpiresAt:    accessToken.ExpiresAt,
				TokenVersion: version,
			}
			_ = s.tokenCache.Set(ctx, sessionID, cached)
		}
	}

	return appContracts.SessionResult{
		SessionID:    sessionID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessToken.ExpiresAt,
	}, nil
}

func (s *SessionService) Refresh(
	ctx context.Context,
	oldRefreshToken string,
	ipAddress string,
	userAgent string,
	deviceID string,
	deviceFingerprint string,
	deviceName string,
) (appContracts.SessionResult, error) {
	now := s.clock()

	userID, sessionID, err := s.tokenGen.ValidateRefresh(oldRefreshToken)
	if err != nil {
		return appContracts.SessionResult{}, application.NewUnauthorized("invalid refresh token")
	}

	oldToken, err := s.refreshTokenRepo.FindByID(ctx, mustParseInt64(sessionID))
	if err != nil {
		return appContracts.SessionResult{}, application.NewUnauthorized("refresh token not found")
	}

	if !oldToken.IsValid(now) {
		return appContracts.SessionResult{}, application.NewUnauthorized("refresh token is expired or revoked")
	}

	newRawToken, err := services.GenerateRandomString(32)
	if err != nil {
		return appContracts.SessionResult{}, fmt.Errorf("generate new session token: %w", err)
	}
	newTokenHash := s.hasher.Hash(newRawToken)

	newRefreshEntity := &entities.RefreshToken{
		UserID:            oldToken.UserID,
		TokenHash:         newTokenHash,
		ExpiresAt:         now.Add(s.policy.RefreshTokenDuration),
		DateCreated:       now,
		DeviceID:          oldToken.DeviceID,
		DeviceName:        oldToken.DeviceName,
		IPAddress:         oldToken.IPAddress,
		UserAgent:         oldToken.UserAgent,
		DeviceFingerprint: oldToken.DeviceFingerprint,
	}

	if deviceID != "" {
		newRefreshEntity.DeviceID = &deviceID
	}
	if deviceName != "" {
		newRefreshEntity.DeviceName = &deviceName
	}
	if ipAddress != "" {
		newRefreshEntity.IPAddress = &ipAddress
	}
	if userAgent != "" {
		newRefreshEntity.UserAgent = &userAgent
	}
	if deviceFingerprint != "" {
		newRefreshEntity.DeviceFingerprint = &deviceFingerprint
	}

	if err := s.refreshTokenRepo.Save(ctx, newRefreshEntity); err != nil {
		return appContracts.SessionResult{}, fmt.Errorf("save new refresh token: %w", err)
	}

	revoked, err := s.refreshTokenRepo.RevokeIfActive(ctx, oldToken.ID, now, &newTokenHash)
	if err != nil {
		return appContracts.SessionResult{}, fmt.Errorf("revoke old refresh token: %w", err)
	}
	if !revoked {
		return appContracts.SessionResult{}, application.NewUnauthorized("refresh token already used")
	}

	newSessionID := fmt.Sprintf("%d", newRefreshEntity.ID)

	accessToken, err := s.tokenGen.GenerateAccess(userID, newSessionID, s.policy.MaxDuration)
	if err != nil {
		return appContracts.SessionResult{}, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.tokenGen.GenerateRefresh(userID, newSessionID, s.policy.RefreshTokenDuration)
	if err != nil {
		return appContracts.SessionResult{}, fmt.Errorf("generate refresh token jwt: %w", err)
	}

	if s.tokenCache != nil {
		cached := &appContracts.CachedToken{
			UserID:       userID,
			SessionID:    newSessionID,
			AccessToken:  accessToken.Value,
			RefreshToken: refreshToken.Value,
			ExpiresAt:    accessToken.ExpiresAt,
		}
		_ = s.tokenCache.Set(ctx, newSessionID, cached)
	}

	return appContracts.SessionResult{
		SessionID:    newSessionID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessToken.ExpiresAt,
	}, nil
}

func mustParseInt64(s string) int64 {
	var v int64
	fmt.Sscanf(s, "%d", &v)
	return v
}

// currentVersion reports the user's session version, or 0 when no version cache
// is configured. An error means the version could not be established, and
// callers must not write a session-cache entry stamped with a guess.
func (s *SessionService) currentVersion(ctx context.Context, userID string) (int16, error) {
	if s.version == nil {
		return 0, nil
	}
	return s.version.GetVersion(ctx, userID)
}
