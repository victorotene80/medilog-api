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

	body, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResult{
		StatusCode: httpRes.StatusCode,
		Headers:    httpRes.Header,
		Body:       body,
	}, nil
}
