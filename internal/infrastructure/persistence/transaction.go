package persistence

import (
	"context"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"gorm.io/gorm"
)

var _ appContracts.TransactionManager = (*TransactionManager)(nil)

type txKey struct{}

// WithTx returns a context carrying tx. Repositories called with that context
// enlist in the transaction instead of taking their own pool connection.
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// conn resolves the handle a repository should use: the ambient transaction when
// the context carries one, otherwise the pool.
//
// Every repository method goes through this rather than r.db directly. That is
// what makes TransactionManager's contract true — a partial rollout would mean
// a write inside a unit of work silently took a second pooled connection,
// committed independently, and survived a rollback of the outer transaction.
// With no transaction on the context it is exactly equivalent to
// r.db.WithContext(ctx), so it is safe as the unconditional default.
func conn(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx)
	}
	return fallback.WithContext(ctx)
}

// TransactionManager runs a unit of work inside one database transaction.
//
// It exists so a state change and the outbox row describing it commit together.
// Without it the outbox write landed after the aggregate's own transaction had
// already committed, on a different connection — so a crash in between lost the
// event silently, which is precisely what the outbox pattern is meant to
// prevent.
type TransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (m *TransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	// Already inside a transaction: join it rather than opening a second one, so
	// nesting stays a single atomic unit.
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
		return fn(ctx)
	}

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(WithTx(ctx, tx))
	})
}
