package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type DefaultHTTPService struct {
	client *http.Client
}

// maxResponseBytes caps how much of an upstream response is buffered.
const maxResponseBytes int64 = 8 << 20 // 8 MiB

func NewDefaultHTTPService(client *http.Client) *DefaultHTTPService {
	if client == nil {
		client = http.DefaultClient
	}

	return &DefaultHTTPService{
		client: client,
	}
}

func (s *DefaultHTTPService) Do(
	ctx context.Context,
	req HTTPRequest,
) (*HTTPResult, error) {
	if strings.TrimSpace(req.URL) == "" {
		return nil, fmt.Errorf("url is required")
	}

	method := req.Method
	if method == "" {
		method = http.MethodGet
	}

	fullURL, err := buildURL(req.URL, req.QueryParams)
	if err != nil {
		return nil, err
	}

	bodyReader, err := buildBody(req.Body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}

	applyHeaders(httpReq, req.Headers, req.Body)

	httpRes, err := s.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpRes.Body.Close()

	// Bounded: this is the single read path for Claude, Twilio, Telnyx and
	// bulksms, the endpoints are env-configurable and the transport honours
	// HTTP_PROXY, so a hijacked or misconfigured upstream streaming a large body
	// would otherwise be buffered in full — limited only by the client timeout,
	// then doubled by the JSON decode. Every current caller parses a small
	// document.
	body, err := io.ReadAll(io.LimitReader(httpRes.Body, maxResponseBytes+1))
	if err != nil {
		return nil, err
	}

	if int64(len(body)) > maxResponseBytes {
		return nil, fmt.Errorf(
			"http service: response from %s exceeds %d bytes", req.URL, maxResponseBytes,
		)
	}

	return &HTTPResult{
		StatusCode: httpRes.StatusCode,
		Headers:    httpRes.Header,
		Body:       body,
	}, nil
}
