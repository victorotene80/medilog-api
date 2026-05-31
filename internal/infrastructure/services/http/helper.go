package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func buildURL(rawURL string, queryParams map[string]string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	q := parsed.Query()
	for key, value := range queryParams {
		q.Set(key, value)
	}

	parsed.RawQuery = q.Encode()
	return parsed.String(), nil
}

func buildBody(body any) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}

	switch v := body.(type) {
	case []byte:
		return bytes.NewReader(v), nil
	case string:
		return strings.NewReader(v), nil
	case io.Reader:
		return v, nil
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(b), nil
	}
}

func applyHeaders(req *http.Request, headers map[string]string, body any) {
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}
}

func decodeJSON[T any](res *HTTPResult) (*T, error) {
	if res == nil {
		return nil, fmt.Errorf("http result is nil")
	}

	var out T

	if len(res.Body) == 0 {
		return &out, nil
	}

	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, err
	}

	return &out, nil
}
