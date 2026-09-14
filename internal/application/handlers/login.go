package handlers

import (
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"

	"context"
	"fmt"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/contracts"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	domainServices "github.com/victorotene80/medilog-api/internal/domain/services"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
	"github.com/victorotene80/medilog-api/internal/shared/requestmeta"
)

type LoginHandler struct {
	userRepo       repository.UserAggregateRepository
	passwordHasher contracts.PasswordHasher
	sessionService appContracts.SessionService
	lockService    *domainServices.AccountLockService
	eventPublisher appContracts.MessagePublisher
	clock          func() time.Time
}

func NewLoginHandler(
	userRepo repository.UserAggregateRepository,
	passwordHasher contracts.PasswordHasher,
	sessionService appContracts.SessionService,
	lockService *domainServices.AccountLockService,
	eventPublisher appContracts.MessagePublisher,
	clock func() time.Time,
) *LoginHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &LoginHandler{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		sessionService: sessionService,
		lockService:    lockService,
		eventPublisher: eventPublisher,
		clock:          clock,
	}
}

func (h *LoginHandler) Handle(
	ctx context.Context,
	cmd command.LoginCommand,
) (*dto.LoginResultDTO, error) {
	now := h.clock()
	meta, _ := requestmeta.FromContext(ctx)

	agg, err := h.resolveUser(ctx, cmd)
	if err != nil || agg == nil {
		return nil, application.NewUnauthorized("invalid credentials")
	}

	if agg.IsLocked(now) {
		return nil, application.NewForbidden("account is locked")
	}

	if agg.User.PasswordHash == nil || !h.passwordHasher.Verify(cmd.Password, *agg.User.PasswordHash) {
		var lockedUntil *time.Time
		if h.lockService.ShouldLock(agg.User.FailedLoginAttempts + 1) {
			t := h.lockService.ComputeLockedUntil(now)
			lockedUntil = &t
		}
		agg.RecordFailedLogin(lockedUntil, now)

		if err := h.userRepo.Update(ctx, agg); err != nil {
			return nil, err
		}
		return nil, application.NewUnauthorized("invalid credentials")
	}

	lastLogin := agg.User.LastLoginAt
	agg.RecordSuccessfulLogin(meta.IPAddress, now)

	if err := h.userRepo.Update(ctx, agg); err != nil {
		return nil, err
	}

	// user.loggedIn is published by UserAggregateRepository.Update, on the same
	// transaction as the row it describes, with the caller's IP, user agent and
	// device taken from the request context. Publishing again here would send an
	// empty batch: the drain has already cleared the aggregate.

	sessionResult, err := h.sessionService.Create(
		ctx,
		fmt.Sprintf("%d", agg.User.ID),
		meta.IPAddress,
		meta.UserAgent,
		meta.DeviceID,
		meta.DeviceFingerprint,
		meta.DeviceName,
	)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResultDTO{
		Status:      "SUCCESS",
		LastLogin:   lastLogin,
		ChallengeID: nil,
		Tokens: contracts.TokenPair{
			AccessToken: contracts.Token{
				Value:     sessionResult.AccessToken.Value,
				ExpiresAt: sessionResult.AccessToken.ExpiresAt,
			},
			RefreshToken: contracts.Token{
				Value:     sessionResult.RefreshToken.Value,
				ExpiresAt: sessionResult.RefreshToken.ExpiresAt,
			},
		},
		OnboardingCompleted: agg.User.IsOnboardingCompleted,
		RequiresOnboarding:  !agg.User.IsOnboardingCompleted,
	}, nil
}

func (h *LoginHandler) resolveUser(
	ctx context.Context,
	cmd command.LoginCommand,
) (*aggregates.UserAggregate, error) {
	if cmd.Email != nil {
		if _, err := valueobjects.NewEmail(*cmd.Email); err != nil {
			return nil, nil
		}
		return h.userRepo.FindByEmail(ctx, *cmd.Email)
	}

	if cmd.Phone != nil {
		if _, err := valueobjects.NewPhone(*cmd.Phone); err != nil {
			return nil, nil
		}
		return h.userRepo.FindByPhone(ctx, *cmd.Phone)
	}

	return nil, nil // neither field set — caller gets "invalid credentials"
}
