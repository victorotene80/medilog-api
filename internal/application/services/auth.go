package services
/*
import (
	"context"
	"time"

	appErrors "github.com/victorotene80/medilog-api/internal/application"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"

	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/cache"

	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	domainContracts "github.com/victorotene80/medilog-api/internal/domain/contracts"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/services/policy"
)

var _ appContracts.AuthService = (*AuthServiceImpl)(nil)

type AuthServiceImpl struct {
	tokenGen      domainContracts.TokenGenerator
	sessionRepo   repository.SessionRepository
	sessionPolicy policy.SessionPolicy
	sessionCache  appContracts.Cache[string, cache.CachedSession]
	auditLogger   appContracts.AuditLogger
}

func NewAuthService(
	tokenGen domainContracts.TokenGenerator,
	sessionRepo repository.SessionRepository,
	sessionPolicy policy.SessionPolicy,
	sessionCache appContracts.Cache[string, cache.CachedSession],
	auditLogger appContracts.AuditLogger,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		tokenGen:      tokenGen,
		sessionRepo:   sessionRepo,
		sessionPolicy: sessionPolicy,
		sessionCache:  sessionCache,
		auditLogger:   auditLogger,
	}
}

func (s *AuthServiceImpl) Authenticate(
	ctx context.Context,
	accessToken string,
) (appContracts.AuthContext, error) {

	userID, sessionID, err := s.tokenGen.ValidateAccess(accessToken)
	if err != nil {
		s.logLoginFailure(ctx, nil, nil, "invalid_token")
		return appContracts.AuthContext{}, appErrors.ErrUnauthenticated
	}

	now := time.Now().UTC()

	var session *aggregates.SessionAggregate

	if s.sessionCache != nil {
		if cached, err := s.sessionCache.Get(ctx, sessionID); err == nil && cached != nil {
			if agg, err := cache.MapCachedToAggregate(*cached); err == nil {
				session = agg
			}
		}
	}

	if session == nil {
		session, err = s.sessionRepo.FindByID(ctx, sessionID)
		if err != nil {
			s.logLoginFailure(ctx, &userID, &sessionID, "session_not_found")
			return appContracts.AuthContext{}, appErrors.ErrUnauthenticated
		}
	}

	if s.sessionPolicy.IsSessionExpired(session.CreatedAt(), session.LastActiveAt(), now) {
		session.Revoke(now, "expired_by_policy")
		_ = s.sessionRepo.Save(ctx, session)

		if s.sessionCache != nil {
			_ = s.sessionCache.Delete(ctx, sessionID)
		}

		s.logLoginFailure(ctx, &userID, &sessionID, "session_expired_by_policy")
		return appContracts.AuthContext{}, appErrors.ErrUnauthenticated
	}

	if !session.IsValid(now) {
		if s.sessionCache != nil {
			_ = s.sessionCache.Delete(ctx, sessionID)
		}
		s.logLoginFailure(ctx, &userID, &sessionID, "session_invalid")
		return appContracts.AuthContext{}, appErrors.ErrUnauthenticated
	}

	session.Touch(now)
	_ = s.sessionRepo.Save(ctx, session)

	if s.sessionCache != nil {
		cached := cache.MapAggregateToCached(session)
		_ = s.sessionCache.Set(ctx, sessionID, &cached)
		_ = s.sessionCache.RefreshTTL(ctx, sessionID)
	}

	s.logLoginSuccess(ctx, userID, sessionID)

	return appContracts.AuthContext{
		UserID:    userID,
		SessionID: sessionID,
	}, nil
}

func (s *AuthServiceImpl) logLoginFailure(
	ctx context.Context,
	userID *string,
	sessionID *string,
	reason string,
) {
	if s.auditLogger == nil {
		return
	}

	rec := dto.AuditRecord{
		Action:        dto.AuditActionLoginFailed,
		UserID:        userID,
		ActorID:       userID,
		SessionID:     sessionID,
		Success:       false,
		FailureReason: &reason,
		Metadata: map[string]any{
			"reason": reason,
		},
	}

	_ = s.auditLogger.Log(ctx, rec)
}

func (s *AuthServiceImpl) logLoginSuccess(
	ctx context.Context,
	userID string,
	sessionID string,
) {
	if s.auditLogger == nil {
		return
	}

	u := userID
	sid := sessionID

	rec := dto.AuditRecord{
		Action:    dto.AuditActionLoginSuccess,
		UserID:    &u,
		ActorID:   &u,
		SessionID: &sid,
		Success:   true,
		Metadata:  map[string]any{},
	}

	_ = s.auditLogger.Log(ctx, rec)
}*/