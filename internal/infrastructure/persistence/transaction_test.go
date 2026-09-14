package persistence

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// A nested Do must join the transaction already on the context rather than open
// a second one — otherwise the "unit of work" spans two connections and the
// atomicity it exists to provide is gone.
func TestTransactionManager_Do_JoinsAmbientTransaction(t *testing.T) {
	ambient := &gorm.DB{}
	ctx := WithTx(context.Background(), ambient)

	// A nil pool proves no new transaction was opened: touching it would panic.
	m := NewTransactionManager(nil)

	called := false
	err := m.Do(ctx, func(inner context.Context) error {
		called = true
		got, ok := inner.Value(txKey{}).(*gorm.DB)
		require.True(t, ok, "inner context must still carry the ambient transaction")
		assert.Same(t, ambient, got)
		return nil
	})

	require.NoError(t, err)
	assert.True(t, called)
}

func TestTransactionManager_Do_PropagatesError(t *testing.T) {
	m := NewTransactionManager(nil)
	sentinel := errors.New("unit of work failed")

	err := m.Do(WithTx(context.Background(), &gorm.DB{}), func(context.Context) error {
		return sentinel
	})

	assert.ErrorIs(t, err, sentinel)
}

// WithTx round-trips the handle that conn later resolves. conn itself dials
// through gorm's WithContext, so exercising its ambient-vs-pool choice needs a
// live handle and belongs in the integration suite.
func TestWithTx_RoundTripsHandle(t *testing.T) {
	tx := &gorm.DB{}

	_, ok := context.Background().Value(txKey{}).(*gorm.DB)
	assert.False(t, ok, "a bare context must not look like it carries a transaction")

	got, ok := WithTx(context.Background(), tx).Value(txKey{}).(*gorm.DB)
	require.True(t, ok)
	assert.Same(t, tx, got)
}
