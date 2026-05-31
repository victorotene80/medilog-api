package contracts

import "context"

type SessionCache interface {
	GetVersion(ctx context.Context, userID string) (int16, error)
	IncrementVersion(ctx context.Context, userID int64) error
}
