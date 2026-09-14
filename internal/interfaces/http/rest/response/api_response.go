package response

import (
	"encoding/json"
	"net/http"
)

// EmptyData is used for Swagger annotations where a named empty struct is required
type EmptyData struct{}

type APIResponse[T any] struct {
	Status    bool   `json:"status"`
	Message   string `json:"message,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Data      *T     `json:"data,omitempty"`
	Errors    any    `json:"errors,omitempty"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func Success[T any](w http.ResponseWriter, statusCode int, code, message string, data *T) {
	resp := APIResponse[T]{
		Status:    true,
		Message:   message,
		RequestID: w.Header().Get("X-Request-ID"),
		Data:      data,
	}
	WriteJSON(w, statusCode, resp)
}

func Error(w http.ResponseWriter, statusCode int, code, message string, errors any) {
	resp := APIResponse[struct{}]{
		Status:    false,
		Message:   message,
		RequestID: w.Header().Get("X-Request-ID"),
		Errors:    errors,
	}
	WriteJSON(w, statusCode, resp)
}
