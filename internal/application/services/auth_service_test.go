package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victorotene80/medilog-api/internal/application"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/services/policy"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
	"github.com/victorotene80/medilog-api/test/testutil"
)

func TestAuthService_CheckAdminAccess_IsAdmin(t *testing.T) {
	mockUserRepo := new(testutil.MockUserRepo)
	ctx := context.Background()

	user := &entities.User{
		ID:                    1,
		Role:                  valueobjects.UserRoleAdmin,
		Status:                valueobjects.UserStatusActive,
		IsOnboardingCompleted: true,
	}

	mockUserRepo.On("FindByID", ctx, int64(1)).Return(user, nil)

	svc := NewAuthService(nil, nil, mockUserRepo, policy.DefaultSessionPolicy(), nil, nil, nil)
	err := svc.CheckAdminAccess(ctx, 1)

	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_CheckAdminAccess_NotAdmin(t *testing.T) {
	mockUserRepo := new(testutil.MockUserRepo)
	ctx := context.Background()

	user := &entities.User{
		ID:                    1,
		Role:                  valueobjects.UserRoleUser,
		Status:                valueobjects.UserStatusActive,
		IsOnboardingCompleted: true,
	}

	mockUserRepo.On("FindByID", ctx, int64(1)).Return(user, nil)

	svc := NewAuthService(nil, nil, mockUserRepo, policy.DefaultSessionPolicy(), nil, nil, nil)
	err := svc.CheckAdminAccess(ctx, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "admin access required")
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_CheckAdminAccess_UserNotFound(t *testing.T) {
	mockUserRepo := new(testutil.MockUserRepo)
	ctx := context.Background()

	mockUserRepo.On("FindByID", ctx, int64(999)).Return(nil, nil)

	svc := NewAuthService(nil, nil, mockUserRepo, policy.DefaultSessionPolicy(), nil, nil, nil)
	err := svc.CheckAdminAccess(ctx, 999)

	assert.Error(t, err)
	assert.Equal(t, application.ErrUserNotFound, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_CheckAdminAccess_AccountLocked(t *testing.T) {
	mockUserRepo := new(testutil.MockUserRepo)
	ctx := context.Background()

	user := &entities.User{
		ID:                    1,
		Role:                  valueobjects.UserRoleAdmin,
		Status:                valueobjects.UserStatusLocked,
		IsOnboardingCompleted: true,
	}

	mockUserRepo.On("FindByID", ctx, int64(1)).Return(user, nil)

	svc := NewAuthService(nil, nil, mockUserRepo, policy.DefaultSessionPolicy(), nil, nil, nil)
	err := svc.CheckAdminAccess(ctx, 1)

	assert.Error(t, err)
	assert.Equal(t, application.ErrAccountLocked, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_CheckAdminAccess_OnboardingNotCompleted(t *testing.T) {
	mockUserRepo := new(testutil.MockUserRepo)
	ctx := context.Background()

	user := &entities.User{
		ID:                    1,
		Role:                  valueobjects.UserRoleAdmin,
		Status:                valueobjects.UserStatusActive,
		IsOnboardingCompleted: false,
	}

	mockUserRepo.On("FindByID", ctx, int64(1)).Return(user, nil)

	svc := NewAuthService(nil, nil, mockUserRepo, policy.DefaultSessionPolicy(), nil, nil, nil)
	err := svc.CheckAdminAccess(ctx, 1)

	assert.Error(t, err)
	assert.Equal(t, application.ErrOnboardingRequired, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_CheckUserAccess_Success(t *testing.T) {
	mockUserRepo := new(testutil.MockUserRepo)
	ctx := context.Background()

	user := &entities.User{
		ID:                    1,
		Status:                valueobjects.UserStatusActive,
		IsOnboardingCompleted: true,
	}

	mockUserRepo.On("FindByID", ctx, int64(1)).Return(user, nil)

	svc := NewAuthService(nil, nil, mockUserRepo, policy.DefaultSessionPolicy(), nil, nil, nil)
	err := svc.CheckUserAccess(ctx, 1)

	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_CheckUserAccess_UserNotFound(t *testing.T) {
	mockUserRepo := new(testutil.MockUserRepo)
	ctx := context.Background()

	mockUserRepo.On("FindByID", ctx, int64(999)).Return(nil, nil)

	svc := NewAuthService(nil, nil, mockUserRepo, policy.DefaultSessionPolicy(), nil, nil, nil)
	err := svc.CheckUserAccess(ctx, 999)

	assert.Error(t, err)
	assert.Equal(t, application.ErrUserNotFound, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_CheckUserAccess_DeletedUser(t *testing.T) {
	mockUserRepo := new(testutil.MockUserRepo)
	ctx := context.Background()

	now := testutil.FixedClock()()
	user := &entities.User{
		ID:        1,
		Status:    valueobjects.UserStatusActive,
		DeletedAt: &now,
	}

	mockUserRepo.On("FindByID", ctx, int64(1)).Return(user, nil)

	svc := NewAuthService(nil, nil, mockUserRepo, policy.DefaultSessionPolicy(), nil, nil, nil)
	err := svc.CheckUserAccess(ctx, 1)

	assert.Error(t, err)
	assert.Equal(t, application.ErrUserNotFound, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Authenticate_Success(t *testing.T) {
	mockTokenGen := new(testutil.MockTokenGenerator)
	mockRefreshRepo := new(testutil.MockRefreshTokenRepo)
	ctx := context.Background()

	refreshToken := &entities.RefreshToken{
		ID:        456,
		UserID:    123,
		ExpiresAt: testutil.FutureTime(7 * 24 * 3600),
	}

	mockTokenGen.On("ValidateAccess", "valid-token").
		Return("123", "456", nil)
	mockRefreshRepo.On("FindByID", ctx, int64(456)).Return(refreshToken, nil)

	svc := NewAuthService(mockTokenGen, mockRefreshRepo, nil, policy.DefaultSessionPolicy(), nil, nil, nil)
	authCtx, err := svc.Authenticate(ctx, "valid-token")

	assert.NoError(t, err)
	assert.Equal(t, "123", authCtx.UserID)
	assert.Equal(t, "456", authCtx.SessionID)
	mockTokenGen.AssertExpectations(t)
}

func TestAuthService_Authenticate_InvalidToken(t *testing.T) {
	mockTokenGen := new(testutil.MockTokenGenerator)
	ctx := context.Background()

	mockTokenGen.On("ValidateAccess", "invalid-token").
		Return("", "", assert.AnError)

	svc := NewAuthService(mockTokenGen, nil, nil, policy.DefaultSessionPolicy(), nil, nil, nil)
	_, err := svc.Authenticate(ctx, "invalid-token")

	assert.Error(t, err)
	assert.Equal(t, application.ErrUnauthenticated, err)
	mockTokenGen.AssertExpectations(t)
}

func TestAuthService_Authenticate_EmptyUserID(t *testing.T) {
	mockTokenGen := new(testutil.MockTokenGenerator)
	ctx := context.Background()

	mockTokenGen.On("ValidateAccess", "token").
		Return("", "123", nil)

	svc := NewAuthService(mockTokenGen, nil, nil, policy.DefaultSessionPolicy(), nil, nil, nil)
	_, err := svc.Authenticate(ctx, "token")

	assert.Error(t, err)
	assert.Equal(t, application.ErrUnauthenticated, err)
	mockTokenGen.AssertExpectations(t)
}

func TestAuthService_CheckUserAccess_VerificationNeeded(t *testing.T) {
	mockUserRepo := new(testutil.MockUserRepo)
	ctx := context.Background()

	user := &entities.User{
		ID:                    1,
		Status:                valueobjects.UserStatusPendingVerification,
		IsOnboardingCompleted: false,
	}

	mockUserRepo.On("FindByID", ctx, int64(1)).Return(user, nil)

	svc := NewAuthService(nil, nil, mockUserRepo, policy.DefaultSessionPolicy(), nil, nil, nil)
	err := svc.CheckUserAccess(ctx, 1)

	assert.Error(t, err)
	assert.Equal(t, application.ErrVerificationNeeded, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_CheckUserAccess_OnboardingRequired(t *testing.T) {
	mockUserRepo := new(testutil.MockUserRepo)
	ctx := context.Background()

	user := &entities.User{
		ID:                    1,
		Status:                valueobjects.UserStatusActive,
		IsOnboardingCompleted: false,
	}

	mockUserRepo.On("FindByID", ctx, int64(1)).Return(user, nil)

	svc := NewAuthService(nil, nil, mockUserRepo, policy.DefaultSessionPolicy(), nil, nil, nil)
	err := svc.CheckUserAccess(ctx, 1)

	assert.Error(t, err)
	assert.Equal(t, application.ErrOnboardingRequired, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Authenticate_WithCache_Expired(t *testing.T) {
	mockTokenGen := new(testutil.MockTokenGenerator)
	mockCache := new(testutil.MockSessionTokenCache)
	ctx := context.Background()

	now := testutil.FixedClock()()
	cachedToken := &appContracts.CachedToken{
		UserID:    "123",
		SessionID: "456",
		ExpiresAt: now.Add(-1 * time.Hour),
	}

	mockTokenGen.On("ValidateAccess", "valid-token").
		Return("123", "456", nil)
	mockCache.On("Get", ctx, "456").Return(cachedToken, nil)

	svc := NewAuthService(mockTokenGen, nil, nil, policy.DefaultSessionPolicy(), mockCache, nil, nil)
	_, err := svc.Authenticate(ctx, "valid-token")

	assert.Error(t, err)
	assert.Equal(t, application.ErrUnauthenticated, err)
	mockTokenGen.AssertExpectations(t)
}

func TestAuthService_Authenticate_WithCache_UserMismatch(t *testing.T) {
	mockTokenGen := new(testutil.MockTokenGenerator)
	mockCache := new(testutil.MockSessionTokenCache)
	ctx := context.Background()

	now := testutil.FixedClock()()
	cachedToken := &appContracts.CachedToken{
		UserID:    "999",
		SessionID: "456",
		ExpiresAt: now.Add(15 * 3600),
	}

	mockTokenGen.On("ValidateAccess", "valid-token").
		Return("123", "456", nil)
	mockCache.On("Get", ctx, "456").Return(cachedToken, nil)

	svc := NewAuthService(mockTokenGen, nil, nil, policy.DefaultSessionPolicy(), mockCache, nil, nil)
	_, err := svc.Authenticate(ctx, "valid-token")

	assert.Error(t, err)
	assert.Equal(t, application.ErrUnauthenticated, err)
	mockTokenGen.AssertExpectations(t)
}

func TestAuthService_Authenticate_RevokedRefreshToken(t *testing.T) {
	mockTokenGen := new(testutil.MockTokenGenerator)
	mockRefreshRepo := new(testutil.MockRefreshTokenRepo)
	ctx := context.Background()

	now := testutil.FixedClock()()
	refreshToken := &entities.RefreshToken{
		ID:        456,
		UserID:    123,
		ExpiresAt: now.Add(7 * 24 * 3600),
		RevokedAt: &now,
	}

	mockTokenGen.On("ValidateAccess", "valid-token").
		Return("123", "456", nil)
	mockRefreshRepo.On("FindByID", ctx, int64(456)).Return(refreshToken, nil)

	svc := NewAuthService(mockTokenGen, mockRefreshRepo, nil, policy.DefaultSessionPolicy(), nil, nil, nil)
	_, err := svc.Authenticate(ctx, "valid-token")

	assert.Error(t, err)
	assert.Equal(t, application.ErrUnauthenticated, err)
	mockTokenGen.AssertExpectations(t)
}

func TestAuthService_Authenticate_ExpiredRefreshToken(t *testing.T) {
	mockTokenGen := new(testutil.MockTokenGenerator)
	mockRefreshRepo := new(testutil.MockRefreshTokenRepo)
	ctx := context.Background()

	refreshToken := &entities.RefreshToken{
		ID:        456,
		UserID:    123,
		ExpiresAt: testutil.PastTime(1 * 24 * 3600),
	}

	mockTokenGen.On("ValidateAccess", "valid-token").
		Return("123", "456", nil)
	mockRefreshRepo.On("FindByID", ctx, int64(456)).Return(refreshToken, nil)

	svc := NewAuthService(mockTokenGen, mockRefreshRepo, nil, policy.DefaultSessionPolicy(), nil, nil, nil)
	_, err := svc.Authenticate(ctx, "valid-token")

	assert.Error(t, err)
	assert.Equal(t, application.ErrUnauthenticated, err)
	mockTokenGen.AssertExpectations(t)
}

func TestAuthService_Authenticate_RefreshTokenNotFound(t *testing.T) {
	mockTokenGen := new(testutil.MockTokenGenerator)
	mockRefreshRepo := new(testutil.MockRefreshTokenRepo)
	ctx := context.Background()

	mockTokenGen.On("ValidateAccess", "valid-token").
		Return("123", "456", nil)
	mockRefreshRepo.On("FindByID", ctx, int64(456)).Return(nil, nil)

	svc := NewAuthService(mockTokenGen, mockRefreshRepo, nil, policy.DefaultSessionPolicy(), nil, nil, nil)
	_, err := svc.Authenticate(ctx, "valid-token")

	assert.Error(t, err)
	assert.Equal(t, application.ErrUnauthenticated, err)
	mockTokenGen.AssertExpectations(t)
}

// A database failure must not be reported as an invalid session. If it were,
// every request during a failover would 401 and clients that treat 401 as
// "session dead" would discard their tokens, pushing the whole user base back
// through SMS OTP. AuthMiddleware maps a non-ErrUnauthenticated error to 500.
func TestAuthService_Authenticate_RepositoryError_IsNotUnauthenticated(t *testing.T) {
	mockTokenGen := new(testutil.MockTokenGenerator)
	mockRefreshRepo := new(testutil.MockRefreshTokenRepo)
	ctx := context.Background()

	dbErr := errors.New("dial tcp 10.0.0.1:5432: connect: connection refused")

	mockTokenGen.On("ValidateAccess", "valid-token").
		Return("123", "456", nil)
	mockRefreshRepo.On("FindByID", ctx, int64(456)).Return(nil, dbErr)

	svc := NewAuthService(mockTokenGen, mockRefreshRepo, nil, policy.DefaultSessionPolicy(), nil, nil, nil)
	_, err := svc.Authenticate(ctx, "valid-token")

	assert.Error(t, err)
	assert.NotErrorIs(t, err, application.ErrUnauthenticated)
	assert.ErrorIs(t, err, dbErr)
	mockTokenGen.AssertExpectations(t)
	mockRefreshRepo.AssertExpectations(t)
}
