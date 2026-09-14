package httperr

import (
	"errors"
	"net/http"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/domain"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

// StatusFrom maps an application/domain error to an HTTP status code.
// Typed AppErrors (application.NewNotFound, NewConflict, ...) take priority;
// remaining sentinels and domain errors are mapped next, and anything unknown
// becomes a 500 so infrastructure bugs are not misreported as client errors.
func StatusFrom(err error) int {
	if err == nil {
		return http.StatusInternalServerError
	}

	var appErr *application.AppError
	if errors.As(err, &appErr) {
		switch appErr.Kind {
		case application.KindNotFound:
			return http.StatusNotFound // 404
		case application.KindConflict:
			return http.StatusConflict // 409
		case application.KindValidation:
			return http.StatusUnprocessableEntity // 422
		case application.KindUnauthorized:
			return http.StatusUnauthorized // 401
		case application.KindForbidden:
			return http.StatusForbidden // 403
		case application.KindRateLimited:
			return http.StatusTooManyRequests // 429
		case application.KindQuotaExceeded:
			// 402, not 429: a spent quota is not a "retry shortly" condition.
			// Clients use the split to show an upgrade prompt rather than a
			// slow-down prompt.
			return http.StatusPaymentRequired // 402
		default:
			return http.StatusInternalServerError
		}
	}

	switch {
	case errors.Is(err, application.ErrUnauthenticated),
		errors.Is(err, application.ErrSessionInvalid):
		return http.StatusUnauthorized // 401
	case errors.Is(err, application.ErrAccountLocked),
		errors.Is(err, application.ErrVerificationNeeded),
		errors.Is(err, application.ErrOnboardingRequired):
		return http.StatusForbidden // 403
	case errors.Is(err, application.ErrUserNotFound),
		errors.Is(err, repository.ErrNotFound):
		// repository.ErrNotFound means a write matched no rows — a concurrent
		// delete, or an id scoped to another user. That is a 404, not the 500
		// this used to fall through to.
		return http.StatusNotFound // 404
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		return http.StatusConflict // 409
	case errors.Is(err, domain.ErrPasswordTooShort),
		errors.Is(err, domain.ErrPasswordMissingUppercase),
		errors.Is(err, domain.ErrPasswordMissingLowercase),
		errors.Is(err, domain.ErrPasswordMissingNumber),
		errors.Is(err, domain.ErrPasswordMissingSpecial):
		return http.StatusUnprocessableEntity // 422
	case errors.Is(err, application.ErrHandlerNotFound),
		errors.Is(err, application.ErrInvalidResult),
		errors.Is(err, application.ErrNilCommand):
		return http.StatusInternalServerError // 500
	default:
		return http.StatusInternalServerError
	}
}
