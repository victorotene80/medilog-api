package handler

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
)

type mockCommandHandler[TCommand messaging.Command, TResult any] struct {
	result TResult
	err    error
}

func (h *mockCommandHandler[TCommand, TResult]) Handle(_ context.Context, _ TCommand) (TResult, error) {
	return h.result, h.err
}

func newMockBus[TCommand messaging.Command, TResult any](result TResult, err error) *messaging.CommandBus {
	bus := messaging.NewCommandBus()
	messaging.MustRegister[TCommand, TResult](bus, &mockCommandHandler[TCommand, TResult]{result: result, err: err})
	return bus
}

func SetupTestContext(userID, sessionID string) context.Context {
	ctx := context.Background()
	authCtx := contracts.AuthContext{
		UserID:    userID,
		SessionID: sessionID,
	}
	return context.WithValue(ctx, contracts.AuthContextKey, authCtx)
}
