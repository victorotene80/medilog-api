package http

import "context"

// "context"
type HTTPService interface {
	Do(ctx context.Context, req HTTPRequest) (*HTTPResult, error)
}
