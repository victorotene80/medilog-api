package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
	"go.uber.org/zap"
)

func PanicRecovery(logger *zap.Logger) func(next http.Handler) http.Handler {
	if logger == nil {
		logger = zap.NewNop()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := string(debug.Stack())
					reqID := r.Header.Get("X-Request-ID")

					logger.Error("panic in HTTP handler",
						zap.Any("panic", rec),
						zap.String("stack", stack),
						zap.String("request_id", reqID),
						zap.String("method", r.Method),
						zap.String("url", r.URL.String()),
						zap.String("remote_addr", r.RemoteAddr),
					)

					response.Error(
						w,
						http.StatusInternalServerError,
						"INTERNAL_SERVER_ERROR",
						"Something went wrong on our side",
						nil,
					)
					//w.Header().Set("Content-Type", "application/json")
					//w.WriteHeader(http.StatusInternalServerError)
					//_, _ = w.Write([]byte(`{"error":"internal_server_error"}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
