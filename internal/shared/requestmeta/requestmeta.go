package requestmeta

import "context"

type key int

const metaKey key = iota

type Meta struct {
	IPAddress         string
	UserAgent         string
	DeviceID          string
	DeviceFingerprint string
	DeviceName        string
	RequestID         string
}

func WithMeta(ctx context.Context, meta Meta) context.Context {
	return context.WithValue(ctx, metaKey, meta)
}

func FromContext(ctx context.Context) (Meta, bool) {
	meta, ok := ctx.Value(metaKey).(Meta)
	return meta, ok
}