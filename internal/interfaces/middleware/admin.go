package middleware

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/victorotene80/medilog-api/internal/application"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
	"go.uber.org/zap"
)

type AdminMiddleware struct {
	authSvc appContracts.AuthService
	logger  *zap.Logger
}

func NewAdminMiddleware(authSvc appContracts.AuthService, logger *zap.Logger) *AdminMiddleware {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &AdminMiddleware{
		authSvc: authSvc,
		logger:  logger,
	}
}

func (m *AdminMiddleware) Handle(next http.Handler) http.Handler {
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

		if err := m.authSvc.CheckAdminAccess(r.Context(), userID); err != nil {
			switch {
			case errors.Is(err, application.ErrUserNotFound):
				response.Error(w, http.StatusUnauthorized, "USER_NOT_FOUND", "User account not found", nil)
			case errors.Is(err, application.ErrAccountLocked):
				response.Error(w, http.StatusForbidden, "ACCOUNT_LOCKED", "Account is locked", nil)
			default:
				m.logger.Error("admin access check failed",
					zap.Error(err),
					zap.String("user_id", authCtx.UserID),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
				)
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "Admin access required", nil)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}
