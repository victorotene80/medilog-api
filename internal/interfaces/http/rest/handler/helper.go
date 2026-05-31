package handler

import (
	"encoding/json"
	"net/http"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
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

func validateRequest(w http.ResponseWriter, v appContracts.Validator, req any) bool {
	if err := v.Struct(req); err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"One or more fields are invalid",
			err.Error(),
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
	if !validateRequest(w, v, req) {
		return nil, false
	}
	return req, true
}

// writeEmptySuccess writes a successful response with an empty data payload.
func writeEmptySuccess(w http.ResponseWriter, statusCode int, code, message string) {
	empty := struct{}{}
	response.Success(w, statusCode, code, message, &empty)
}
