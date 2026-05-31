package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

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
	clock func() time.Time
}

func NewSessionService(
	refreshTokenRepo repository.RefreshTokenRepository,
	tokenGen domainContracts.TokenGenerator,
	hasher *services.SessionKeyHasher,
	policy policy.SessionPolicy,
	tokenCache appContracts.Cache[string, appContracts.CachedToken],
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
		clock: clock,
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
		cached := &appContracts.CachedToken{
			UserID:       userID,
			SessionID:    sessionID,
			AccessToken:  accessToken.Value,
			RefreshToken: refreshToken.Value,
			ExpiresAt:    accessToken.ExpiresAt,
		}
		// cache key is sessionID — lets auth middleware do O(1) lookups
		// ignore cache errors; a cache miss falls back to DB validation
		_ = s.tokenCache.Set(ctx, sessionID, cached)
	}

	return appContracts.SessionResult{
		SessionID:    sessionID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessToken.ExpiresAt,
	}, nil
}
