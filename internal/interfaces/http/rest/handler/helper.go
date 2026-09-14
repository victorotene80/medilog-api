package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/victorotene80/medilog-api/internal/application"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
	"go.uber.org/zap"
)

func decodeJSON[T any](w http.ResponseWriter, r *http.Request) (*T, bool) {
	defer r.Body.Close()

	req := new(T)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(req); err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST_BODY",
			"Invalid JSON payload",
			nil,
		)
		return nil, false
	}

	return req, true
}

func validateRequest(w http.ResponseWriter, r *http.Request, v appContracts.Validator, req any) bool {
	if err := v.Struct(req); err != nil {
		zap.L().Warn("request validation failed", zap.Error(err))
		response.Error(
			w,
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"One or more fields are invalid",
			nil,
		)
		return false
	}
	return true
}

func decodeAndValidate[T any](
	w http.ResponseWriter,
	r *http.Request,
	v appContracts.Validator,
) (*T, bool) {
	req, ok := decodeJSON[T](w, r)
	if !ok {
		return nil, false
	}
	if !validateRequest(w, r, v, req) {
		return nil, false
	}
	return req, true
}

// publicIDParam reads a {publicId}-style path parameter and rejects anything
// that is not a well-formed UUID, writing a 400 and returning false.
//
// The format check is not cosmetic: these columns are Postgres `uuid`, so
// passing arbitrary text straight through to a WHERE clause raises
// "invalid input syntax for type uuid" and surfaces as a 500. Validating here
// keeps a malformed id a 400 and leaves 404 to mean "well-formed but unknown".
func publicIDParam(w http.ResponseWriter, r *http.Request, name, message string) (string, bool) {
	value := strings.TrimSpace(chi.URLParam(r, name))
	if value == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", message, nil)
		return "", false
	}

	if _, err := uuid.Parse(value); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", message, nil)
		return "", false
	}

	return value, true
}

// writeEmptySuccess writes a successful response with an empty data payload.
func writeEmptySuccess(w http.ResponseWriter, statusCode int, code, message string) {
	empty := struct{}{}
	response.Success(w, statusCode, code, message, &empty)
}

func logAndRespond(w http.ResponseWriter, statusCode int, code, message string, err error) {
	if err != nil {
		zap.L().Error(message, zap.String("code", code), zap.Error(err))

		// Surface the underlying text only for deliberate application errors,
		// whose messages are authored for end users ("emergency contact already
		// exists"). Anything else is a wrapped infrastructure error whose text
		// leaks driver, query or schema internals, so the caller's curated
		// message stands instead.
		var appErr *application.AppError
		if statusCode >= 400 && statusCode < 500 && errors.As(err, &appErr) {
			message = appErr.Message
		}
	}
	response.Error(w, statusCode, code, message, nil)
}
