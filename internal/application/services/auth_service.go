package services

import (
	"context"
	"strconv"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	domainContracts "github.com/victorotene80/medilog-api/internal/domain/contracts"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/services/policy"
)

var _ appContracts.AuthService = (*authService)(nil)

type authService struct {
	tokenGen     domainContracts.TokenGenerator
	refreshRepo  repository.RefreshTokenRepository
	userRepo     repository.UserRepository
	policy       policy.SessionPolicy
	sessionCache appContracts.Cache[string, appContracts.CachedToken]
	clock        func() time.Time
	version      appContracts.SessionCache
}

func NewAuthService(
	tokenGen domainContracts.TokenGenerator,
	refreshRepo repository.RefreshTokenRepository,
	userRepo repository.UserRepository,
	policy policy.SessionPolicy,
	sessionCache appContracts.Cache[string, appContracts.CachedToken],
	clock func() time.Time,
	version appContracts.SessionCache,
) appContracts.AuthService {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}
	return &authService{
		tokenGen:     tokenGen,
		refreshRepo:  refreshRepo,
		userRepo:     userRepo,
		policy:       policy,
		sessionCache: sessionCache,
		clock:        clock,
		version:      version,
	}
}

func (s *authService) Authenticate(
	ctx context.Context,
	accessToken string,
) (appContracts.AuthContext, error) {
	userID, sessionID, err := s.tokenGen.ValidateAccess(accessToken)
	if err != nil || userID == "" || sessionID == "" {
		return appContracts.AuthContext{}, application.ErrUnauthenticated
	}

	now := s.clock()

	if s.sessionCache != nil {
		if cached, err := s.sessionCache.Get(ctx, sessionID); err == nil && cached != nil {
			if cached.ExpiresAt.Before(now) || cached.UserID != userID {
				return appContracts.AuthContext{}, application.ErrUnauthenticated
			}

			if s.version != nil {
				currentVersion, err := s.version.GetVersion(ctx, userID)
				if err == nil && cached.TokenVersion != currentVersion {
					goto slowPath
				}
			}

			return appContracts.AuthContext{UserID: userID, SessionID: sessionID}, nil
		}
	}

slowPath:
	rtID, err := strconv.ParseInt(sessionID, 10, 64)
	if err != nil {
		return appContracts.AuthContext{}, application.ErrUnauthenticated
	}

	rt, err := s.refreshRepo.FindByID(ctx, rtID)
	if err != nil || rt == nil {
		return appContracts.AuthContext{}, application.ErrUnauthenticated
	}

	if rt.IsRevoked() || rt.IsExpired(now) {
		return appContracts.AuthContext{}, application.ErrUnauthenticated
	}

	userIDStr := strconv.FormatInt(rt.UserID, 10)
	if userIDStr != userID {
		return appContracts.AuthContext{}, application.ErrUnauthenticated
	}

	if s.sessionCache != nil {
		_ = s.sessionCache.Set(ctx, sessionID, &appContracts.CachedToken{
			UserID:    userID,
			SessionID: sessionID,
			ExpiresAt: rt.ExpiresAt,
		})
	}

	return appContracts.AuthContext{UserID: userID, SessionID: sessionID}, nil
}

func (s *authService) CheckUserAccess(ctx context.Context, userID int64) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if user == nil || user.IsDeleted() {
		return application.ErrUserNotFound
	}

	if user.Status.IsLocked() {
		return application.ErrAccountLocked
	}

	if user.Status.IsPendingVerification() {
		return application.ErrVerificationNeeded
	}

	if !user.IsOnboardingCompleted {
		return application.ErrOnboardingRequired
	}

	return nil
}
