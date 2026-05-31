package http

import "net/http"

type HTTPRequest struct {
	Method      string
	URL         string
	Headers     map[string]string
	QueryParams map[string]string
	Body        any
}

type HTTPResult struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}
