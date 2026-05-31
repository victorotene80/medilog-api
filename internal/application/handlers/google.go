package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/contracts"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/shared/requestmeta"
)

type GoogleAuthHandler struct {
	userRepo         repository.UserRepository
	userAggregate    repository.UserAggregateRepository
	authProviderRepo repository.UserAuthProviderRepository
	googleService    appContracts.GoogleAuthService
	sessionService   appContracts.SessionService
	eventPublisher   appContracts.MessagePublisher
	clock            func() time.Time
}

func NewGoogleAuthHandler(
	userRepo repository.UserRepository,
	userAggregate repository.UserAggregateRepository,
	authProviderRepo repository.UserAuthProviderRepository,
	googleService appContracts.GoogleAuthService,
	sessionService appContracts.SessionService,
	eventPublisher appContracts.MessagePublisher,
	clock func() time.Time,
) *GoogleAuthHandler {
	return &GoogleAuthHandler{
		userRepo:         userRepo,
		userAggregate:    userAggregate,
		authProviderRepo: authProviderRepo,
		googleService:    googleService,
		sessionService:   sessionService,
		eventPublisher:   eventPublisher,
		clock:            clock,
	}
}

func (h *GoogleAuthHandler) Handle(
	ctx context.Context,
	cmd command.GoogleLoginCommand,
) (*dto.GoogleAuthResultDTO, error) {
	now := h.clock().UTC()

	meta, _ := requestmeta.FromContext(ctx)

	googleUser, err := h.googleService.VerifyIDToken(ctx, cmd.IDToken)
	if err != nil {
		return nil, err
	}

	email := strings.TrimSpace(googleUser.Email)
	if email == "" {
		return nil, fmt.Errorf("google account did not return an email")
	}

	googleID := strings.TrimSpace(googleUser.GoogleID)
	if googleID == "" {
		return nil, fmt.Errorf("google account did not return a provider id")
	}

	firstName, lastName := resolveGoogleName(
		googleUser.FirstName,
		googleUser.LastName,
		googleUser.Name,
	)

	// Case 1:
	// Google provider is already linked to an existing user.
	existingProvider, err := h.authProviderRepo.FindByProviderUID(
		ctx,
		"google",
		googleID,
	)
	if err != nil {
		return nil, err
	}

	if existingProvider != nil {
		user, err := h.userRepo.FindByID(ctx, existingProvider.UserID)
		if err != nil {
			return nil, err
		}

		if user == nil {
			return nil, fmt.Errorf("user not found for google provider")
		}

		return h.loginExistingUser(ctx, user, googleUser.PictureURL, meta, now)
	}

	// Case 2:
	// A normal email/password user already exists with this email,
	// but Google has not been linked yet.
	existingUser, err := h.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		provider := entities.NewUserAuthProvider(
			existingUser.ID,
			"google",
			googleID,
			&email,
			false,
		)

		if err := h.authProviderRepo.Create(ctx, provider); err != nil {
			return nil, err
		}

		h.publishEvent(
			ctx,
			"auth.user.google-linked.v1",
			"google-link",
			existingUser.ID,
			meta,
		)

		return h.loginExistingUser(ctx, existingUser, googleUser.PictureURL, meta, now)
	}

	// Case 3:
	// Brand-new Google user.
	// Persist immediately, keep onboarding incomplete, then return tokens.
	user := entities.NewGoogleUser(
		email,
		firstName,
		lastName,
		googleUser.PictureURL,
		now,
	)

	agg := aggregates.NewUserAggregate(user)

	if err := h.userAggregate.Save(ctx, agg); err != nil {
		return nil, err
	}

	provider := entities.NewUserAuthProvider(
		agg.User.ID,
		"google",
		googleID,
		&email,
		true,
	)

	if err := h.authProviderRepo.Create(ctx, provider); err != nil {
		return nil, err
	}

	h.publishEvent(
		ctx,
		"auth.user.google-created.v1",
		"google-register",
		agg.User.ID,
		meta,
	)

	return h.loginExistingUser(ctx, agg.User, googleUser.PictureURL, meta, now)
}

func (h *GoogleAuthHandler) loginExistingUser(
	ctx context.Context,
	user *entities.User,
	pictureURL string,
	meta requestmeta.Meta,
	now time.Time,
) (*dto.GoogleAuthResultDTO, error) {
	if user == nil {
		return nil, fmt.Errorf("user is required")
	}

	user.LastLoginAt = &now
	user.LastActiveAt = &now
	user.LastLoginIP = &meta.IPAddress
	user.ResetFailedLogins()
	user.UpdatedAt = now

	if pictureURL != "" && user.AvatarURL == nil {
		user.AvatarURL = &pictureURL
	}

	if err := h.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	tokens, err := h.createSession(ctx, user.ID, meta)
	if err != nil {
		return nil, err
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	return &dto.GoogleAuthResultDTO{
		Status:              dto.GoogleAuthStatusLogin,
		UserID:              user.PublicID,
		Email:               email,
		FirstName:           user.FirstName,
		LastName:            user.LastName,
		LastLogin:           user.LastLoginAt,
		Tokens:              tokens,
		OnboardingCompleted: user.IsOnboardingCompleted,
		RequiresOnboarding:  !user.IsOnboardingCompleted,
	}, nil
}

func (h *GoogleAuthHandler) createSession(
	ctx context.Context,
	userID int64,
	meta requestmeta.Meta,
) (*contracts.TokenPair, error) {
	sessionResult, err := h.sessionService.Create(
		ctx,
		fmt.Sprintf("%d", userID),
		meta.IPAddress,
		meta.UserAgent,
		meta.DeviceID,
		meta.DeviceFingerprint,
		meta.DeviceName,
	)
	if err != nil {
		return nil, err
	}

	return &contracts.TokenPair{
		AccessToken: contracts.Token{
			Value:     sessionResult.AccessToken.Value,
			ExpiresAt: sessionResult.AccessToken.ExpiresAt,
		},
		RefreshToken: contracts.Token{
			Value:     sessionResult.RefreshToken.Value,
			ExpiresAt: sessionResult.RefreshToken.ExpiresAt,
		},
	}, nil
}

func (h *GoogleAuthHandler) publishEvent(
	ctx context.Context,
	name string,
	action string,
	userID int64,
	meta requestmeta.Meta,
) {
	if h.eventPublisher == nil {
		return
	}

	eventMeta := messaging.Context{
		Kind:          messaging.KindIntegrationEvent,
		Name:          name,
		AggregateType: "user",
		Action:        action,
		IPAddress:     &meta.IPAddress,
		UserAgent:     &meta.UserAgent,
		DeviceID:      &meta.DeviceID,
	}

	_ = h.eventPublisher.Publish(ctx, nil, eventMeta.ToMetadata())
}

func resolveGoogleName(first string, last string, full string) (string, string) {
	first = strings.TrimSpace(first)
	last = strings.TrimSpace(last)

	if first != "" || last != "" {
		return first, last
	}

	full = strings.TrimSpace(full)
	if full == "" {
		return "User", ""
	}

	parts := strings.SplitN(full, " ", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return parts[0], ""
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
