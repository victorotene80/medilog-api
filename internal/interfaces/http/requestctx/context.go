package requestctx

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/shared/requestmeta"
)

type RequestMeta = requestmeta.Meta

func WithMeta(ctx context.Context, meta RequestMeta) context.Context {
	return requestmeta.WithMeta(ctx, meta)
}

func MetaFrom(ctx context.Context) (RequestMeta, bool) {
	return requestmeta.FromContext(ctx)
}
