package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victorotene80/medilog-api/internal/application"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"go.uber.org/zap"
)

func TestPanicRecovery_NoPanic(t *testing.T) {
	logger := zap.NewNop()
	mw := PanicRecovery(logger)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPanicRecovery_RecoverPanic(t *testing.T) {
	logger := zap.NewNop()
	mw := PanicRecovery(logger)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAdminMiddleware_AllowsAdmin(t *testing.T) {
	authSvc := &mockAuthService{checkAdminAccessResult: nil}
	logger := zap.NewNop()
	mw := NewAdminMiddleware(authSvc, logger)

	called := false
	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := req.Context()
	authCtx := appContracts.AuthContext{UserID: "1", SessionID: "100"}
	ctx = context.WithValue(ctx, appContracts.AuthContextKey, authCtx)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminMiddleware_RejectsNonAdmin(t *testing.T) {
	authSvc := &mockAuthService{checkAdminAccessResult: application.NewForbidden("admin access required")}
	logger := zap.NewNop()
	mw := NewAdminMiddleware(authSvc, logger)

	called := false
	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := req.Context()
	authCtx := appContracts.AuthContext{UserID: "1", SessionID: "100"}
	ctx = context.WithValue(ctx, appContracts.AuthContextKey, authCtx)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminMiddleware_NoAuthContext(t *testing.T) {
	authSvc := &mockAuthService{}
	logger := zap.NewNop()
	mw := NewAdminMiddleware(authSvc, logger)

	called := false
	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminMiddleware_UserNotFound(t *testing.T) {
	authSvc := &mockAuthService{checkAdminAccessResult: application.ErrUserNotFound}
	logger := zap.NewNop()
	mw := NewAdminMiddleware(authSvc, logger)

	called := false
	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := req.Context()
	authCtx := appContracts.AuthContext{UserID: "1", SessionID: "100"}
	ctx = context.WithValue(ctx, appContracts.AuthContextKey, authCtx)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminMiddleware_AccountLocked(t *testing.T) {
	authSvc := &mockAuthService{checkAdminAccessResult: application.ErrAccountLocked}
	logger := zap.NewNop()
	mw := NewAdminMiddleware(authSvc, logger)

	called := false
	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := req.Context()
	authCtx := appContracts.AuthContext{UserID: "1", SessionID: "100"}
	ctx = context.WithValue(ctx, appContracts.AuthContextKey, authCtx)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	authCtx := appContracts.AuthContext{UserID: "1", SessionID: "100"}
	authSvc := &mockAuthService{authenticateResult: authCtx, authenticateErr: nil}
	logger := zap.NewNop()
	mw := NewAuthMiddleware(authSvc, logger)

	called := false
	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	authSvc := &mockAuthService{authenticateErr: application.ErrUnauthenticated}
	logger := zap.NewNop()
	mw := NewAuthMiddleware(authSvc, logger)

	called := false
	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	authSvc := &mockAuthService{}
	logger := zap.NewNop()
	mw := NewAuthMiddleware(authSvc, logger)

	called := false
	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	authSvc := &mockAuthService{}
	logger := zap.NewNop()
	mw := NewAuthMiddleware(authSvc, logger)

	called := false
	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token some-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_EmptyBearer(t *testing.T) {
	authSvc := &mockAuthService{}
	logger := zap.NewNop()
	mw := NewAuthMiddleware(authSvc, logger)

	called := false
	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_RequireOnboardingCompleted_Success(t *testing.T) {
	authSvc := &mockAuthService{checkUserAccessResult: nil}
	logger := zap.NewNop()
	mw := NewAuthMiddleware(authSvc, logger)

	called := false
	handler := mw.RequireOnboardingCompleted(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	ctx := req.Context()
	authCtx := appContracts.AuthContext{UserID: "1", SessionID: "100"}
	ctx = context.WithValue(ctx, appContracts.AuthContextKey, authCtx)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_RequireOnboardingCompleted_OnboardingRequired(t *testing.T) {
	authSvc := &mockAuthService{checkUserAccessResult: application.ErrOnboardingRequired}
	logger := zap.NewNop()
	mw := NewAuthMiddleware(authSvc, logger)

	called := false
	handler := mw.RequireOnboardingCompleted(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	ctx := req.Context()
	authCtx := appContracts.AuthContext{UserID: "1", SessionID: "100"}
	ctx = context.WithValue(ctx, appContracts.AuthContextKey, authCtx)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuthMiddleware_RequireOnboardingCompleted_NoAuthContext(t *testing.T) {
	authSvc := &mockAuthService{}
	logger := zap.NewNop()
	mw := NewAuthMiddleware(authSvc, logger)

	called := false
	handler := mw.RequireOnboardingCompleted(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_RequireOnboardingCompleted_VerificationNeeded(t *testing.T) {
	authSvc := &mockAuthService{checkUserAccessResult: application.ErrVerificationNeeded}
	logger := zap.NewNop()
	mw := NewAuthMiddleware(authSvc, logger)

	called := false
	handler := mw.RequireOnboardingCompleted(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	ctx := req.Context()
	authCtx := appContracts.AuthContext{UserID: "1", SessionID: "100"}
	ctx = context.WithValue(ctx, appContracts.AuthContextKey, authCtx)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
