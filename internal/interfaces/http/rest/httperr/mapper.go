package httperr

import (
	"errors"
	"net/http"

	"github.com/victorotene80/medilog-api/internal/domain"
	appErrors "github.com/victorotene80/medilog-api/internal/application"
)

func StatusFrom(err error) int {
	switch {
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		return http.StatusConflict // 409
	case errors.Is(err, domain.ErrPasswordTooShort),
		errors.Is(err, domain.ErrPasswordMissingUppercase),
		errors.Is(err, domain.ErrPasswordMissingLowercase),
		errors.Is(err, domain.ErrPasswordMissingNumber),
		errors.Is(err, domain.ErrPasswordMissingSpecial):
		return http.StatusUnprocessableEntity // 422
	case errors.Is(err, appErrors.ErrHandlerNotFound):
		return http.StatusInternalServerError // 500
	default:
		return http.StatusInternalServerError
	}
}