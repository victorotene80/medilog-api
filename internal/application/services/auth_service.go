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

			if s.version == nil {
				return appContracts.AuthContext{UserID: userID, SessionID: sessionID}, nil
			}

			// Fail closed. This comparison is the only thing standing between a
			// revoked session and a valid-looking cache entry, so a version we
			// cannot read must fall through to the database rather than be
			// treated as a match.
			if currentVersion, err := s.version.GetVersion(ctx, userID); err == nil &&
				cached.TokenVersion == currentVersion {
				return appContracts.AuthContext{UserID: userID, SessionID: sessionID}, nil
			}
		}
	}

	rtID, err := strconv.ParseInt(sessionID, 10, 64)
	if err != nil {
		return appContracts.AuthContext{}, application.ErrUnauthenticated
	}

	rt, err := s.refreshRepo.FindByID(ctx, rtID)
	if err != nil {
		// Not ErrUnauthenticated: this is the database being unreachable, not the
		// session being invalid. Collapsing the two 401s every request during a
		// failover, and clients that read 401 as "session dead" would discard
		// their tokens and push the entire user base back through SMS OTP. The
		// cache entry is deliberately left intact so recovery is not cold.
		// AuthMiddleware maps a non-ErrUnauthenticated error to 500.
		return appContracts.AuthContext{}, fmt.Errorf("find session: %w", err)
	}

	if rt == nil {
		if s.sessionCache != nil {
			_ = s.sessionCache.Delete(ctx, sessionID)
		}
		return appContracts.AuthContext{}, application.ErrUnauthenticated
	}

	if rt.IsRevoked() || rt.IsExpired(now) {
		if s.sessionCache != nil {
			_ = s.sessionCache.Delete(ctx, sessionID)
		}
		return appContracts.AuthContext{}, application.ErrUnauthenticated
	}

	userIDStr := strconv.FormatInt(rt.UserID, 10)
	if userIDStr != userID {
		if s.sessionCache != nil {
			_ = s.sessionCache.Delete(ctx, sessionID)
		}
		return appContracts.AuthContext{}, application.ErrUnauthenticated
	}

	if s.sessionCache != nil && s.version != nil {
		// Only cache an entry we can stamp with the current version; an unstamped
		// entry would read as version 0 and survive a later revocation.
		if version, err := s.version.GetVersion(ctx, userID); err == nil {
			_ = s.sessionCache.Set(ctx, sessionID, &appContracts.CachedToken{
				UserID:       userID,
				SessionID:    sessionID,
				ExpiresAt:    rt.ExpiresAt,
				TokenVersion: version,
			})
		}
	}

	return appContracts.AuthContext{UserID: userID, SessionID: sessionID}, nil
}

// checkUser applies the account gates shared by every authenticated route to an
// already-loaded user. Splitting the predicate from the load is what lets the
// three exported checks stay identical without any of them paying for a second
// SELECT: CheckUserAccess used to call CheckUserActive and then re-load the same
// row for one boolean, doubling the user query on the entire full-app surface.
func checkUser(user *entities.User, requireOnboarding bool) error {
	if user == nil || user.IsDeleted() {
		return application.ErrUserNotFound
	}

	if user.Status.IsLocked() {
		return application.ErrAccountLocked
	}

	if user.Status.IsPendingVerification() {
		return application.ErrVerificationNeeded
	}

	if requireOnboarding && !user.IsOnboardingCompleted {
		return application.ErrOnboardingRequired
	}

	return nil
}

// CheckUserActive verifies the account exists and is usable, without requiring
// onboarding to be finished. Endpoints that a user must reach *in order to*
// complete onboarding (creating the first emergency contact) gate on this.
func (s *authService) CheckUserActive(ctx context.Context, userID int64) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	return checkUser(user, false)
}

// CheckUserAccess is CheckUserActive plus the onboarding requirement.
func (s *authService) CheckUserAccess(ctx context.Context, userID int64) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	return checkUser(user, true)
}

func (s *authService) CheckAdminAccess(ctx context.Context, userID int64) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := checkUser(user, true); err != nil {
		return err
	}

	if !user.Role.IsAdmin() {
		return application.NewForbidden("admin access required")
	}

	return nil
}
