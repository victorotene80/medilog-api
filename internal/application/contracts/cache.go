package contracts

import (
	"context"
)

type Cache[K comparable, V any] interface {
	Get(ctx context.Context, key K) (*V, error)
	Set(ctx context.Context, key K, value *V) error
	Delete(ctx context.Context, key K) error
	RefreshTTL(ctx context.Context, key K) error
}
