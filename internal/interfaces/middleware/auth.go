package middleware

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/victorotene80/medilog-api/internal/application"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
	"go.uber.org/zap"
)

type AuthMiddleware struct {
	authSvc appContracts.AuthService
	logger  *zap.Logger
}

func NewAuthMiddleware(authSvc appContracts.AuthService, logger *zap.Logger) *AuthMiddleware {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &AuthMiddleware{
		authSvc: authSvc,
		logger:  logger,
	}
}

func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header missing or malformed", nil)
			return
		}

		authCtx, err := m.authSvc.Authenticate(r.Context(), token)
		if err != nil {
			if errors.Is(err, application.ErrUnauthenticated) {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token", nil)
				return
			}

			m.logger.Error("authentication infrastructure error",
				zap.Error(err),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)
			response.Error(w, http.StatusInternalServerError, "AUTH_ERROR", "Authentication failed", nil)
			return
		}

		ctx := context.WithValue(r.Context(), appContracts.AuthContextKey, authCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth attaches the auth context when the caller presents a valid
// bearer token, and otherwise lets the request through anonymously. It exists
// for endpoints that accept anonymous callers but should still attribute the
// request when a signed-in user makes it. A token that is present but invalid
// is still rejected — presenting credentials that do not validate is an error,
// not an anonymous request.
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	authenticated := m.Handle(next)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The only difference from Handle is what a *missing* token means. Once a
		// token is present the two must validate, classify and log identically,
		// so this delegates rather than repeating the body — the copy previously
		// here meant every fix to the error handling had to be made twice.
		if _, ok := bearerToken(r); !ok {
			next.ServeHTTP(w, r)
			return
		}

		authenticated.ServeHTTP(w, r)
	})
}

// RequireOnboardingCompleted rejects callers who have not finished onboarding.
// It is the default gate for authenticated routes.
func (m *AuthMiddleware) RequireOnboardingCompleted(next http.Handler) http.Handler {
	return m.requireUserCheck(next, m.authSvc.CheckUserAccess)
}

// RequireActiveUser applies every account check except onboarding. It exists
// for the endpoints a user must reach in order to complete onboarding — gating
// those on RequireOnboardingCompleted would lock a new user out of the only
// route that can finish their onboarding.
func (m *AuthMiddleware) RequireActiveUser(next http.Handler) http.Handler {
	return m.requireUserCheck(next, m.authSvc.CheckUserActive)
}

func (m *AuthMiddleware) requireUserCheck(
	next http.Handler,
	check func(context.Context, int64) error,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authCtx, ok := r.Context().Value(appContracts.AuthContextKey).(appContracts.AuthContext)
		if !ok {
			response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication context missing", nil)
			return
		}

		userID, err := strconv.ParseInt(authCtx.UserID, 10, 64)
		if err != nil || userID <= 0 {
			response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authenticated user", nil)
			return
		}

		if err := check(r.Context(), userID); err != nil {
			switch {
			case errors.Is(err, application.ErrUserNotFound):
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "User account not found", nil)
			case errors.Is(err, application.ErrAccountLocked):
				response.Error(w, http.StatusForbidden, "ACCOUNT_NOT_ALLOWED", "Your account is not allowed to access this resource", nil)
			case errors.Is(err, application.ErrVerificationNeeded):
				response.Error(w, http.StatusForbidden, "VERIFICATION_REQUIRED", "Please verify your account before continuing.", map[string]any{
					"requires_verification": true,
					"next_step":             "otp_verification",
				})
			case errors.Is(err, application.ErrOnboardingRequired):
				response.Error(w, http.StatusForbidden, "ONBOARDING_REQUIRED", "Please complete onboarding before continuing.", map[string]any{
					"requires_onboarding": true,
					"onboarding_step":     "emergency_contact",
				})
			default:
				m.logger.Error("user access check failed",
					zap.Error(err),
					zap.String("user_id", authCtx.UserID),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
				)
				response.Error(w, http.StatusInternalServerError, "ONBOARDING_CHECK_FAILED", "Could not verify onboarding status", nil)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", false
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}

	return token, true
}
