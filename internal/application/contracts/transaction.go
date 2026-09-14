package contracts

import "context"

// TransactionManager runs a unit of work atomically.
//
// The context passed to fn carries the transaction; repositories invoked with it
// enlist in that transaction automatically, so handlers compose a unit of work
// without any database type crossing into the application layer.
type TransactionManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
