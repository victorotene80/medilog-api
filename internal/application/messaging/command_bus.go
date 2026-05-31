package messaging

import (
	"context"
	"fmt"
	"sync"

	"github.com/victorotene80/medilog-api/internal/application"
)

type CommandBus struct {
	mu         sync.RWMutex
	handlers   map[string]any
	middleware []Middleware
}

func NewCommandBus() *CommandBus {
	return &CommandBus{
		handlers:   make(map[string]any),
		middleware: []Middleware{},
	}
}

func Register[TCommand Command, TResult any](
	bus *CommandBus,
	handler CommandHandler[TCommand, TResult],
) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	key := getTypeKey[TCommand]()
	fmt.Println("REGISTER key:", key, "bus:", bus) // TEMP debug

	if _, exists := bus.handlers[key]; exists {
		return application.ErrHandlerExists
	}

	bus.handlers[key] = handler
	return nil
}

func MustRegister[TCommand Command, TResult any](
	bus *CommandBus,
	handler CommandHandler[TCommand, TResult],
) {
	if err := Register(bus, handler); err != nil {
		panic(err)
	}
}

// NOTE: simplified constraint here: TCommand must satisfy Command
func Execute[TCommand Command, TResult any](
	bus *CommandBus,
	ctx context.Context,
	cmd TCommand,
) (TResult, error) {
	var zero TResult
	if any(cmd) == nil {
		return zero, application.ErrNilCommand
	}

	key := getTypeKey[TCommand]()
	fmt.Println("EXECUTE key:", key, "bus:", bus) // TEMP debug

	bus.mu.RLock()
	rawHandler, exists := bus.handlers[key]
	bus.mu.RUnlock()

	if !exists {
		return zero, application.ErrHandlerNotFound
	}

	handler, ok := rawHandler.(CommandHandler[TCommand, TResult])
	if !ok {
		// This is the "TResult mismatch" case
		return zero, application.ErrHandlerNotFound
	}

	finalHandler := func(ctx context.Context, c Command) (any, error) {
		return handler.Handle(ctx, c.(TCommand))
	}

	// apply middleware
	for i := len(bus.middleware) - 1; i >= 0; i-- {
		finalHandler = bus.middleware[i](finalHandler)
	}

	result, err := finalHandler(ctx, cmd)
	if err != nil {
		return zero, err
	}

	typedResult, ok := result.(TResult)
	if !ok {
		return zero, application.ErrInvalidResult
	}

	return typedResult, nil
}

func (bus *CommandBus) Use(mw Middleware) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.middleware = append(bus.middleware, mw)
}

func getTypeKey[TCommand Command]() string {
	var t TCommand
	return fmt.Sprintf("%T", t)
}

/*func Register[TCommand Command, TResult any](
	bus *CommandBus,
	handler CommandHandler[TCommand, TResult],
) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	key := getTypeKey[TCommand]()
	if _, exists := bus.handlers[key]; exists {
		return application.ErrHandlerExists
	}

	bus.handlers[key] = handler
	return nil
}

func MustRegister[TCommand Command, TResult any](
	bus *CommandBus,
	handler CommandHandler[TCommand, TResult],
) {
	if err := Register(bus, handler); err != nil {
		panic(err)
	}
}

func Execute[TCommand Command, TResult any](
    bus *CommandBus,
    ctx context.Context,
    cmd TCommand,
) (TResult, error) {
    var zero TResult
    if any(cmd) == nil {
        return zero, application.ErrNilCommand
    }

    key := getTypeKey[TCommand]()

    bus.mu.RLock()
    rawHandler, exists := bus.handlers[key]
    bus.mu.RUnlock()

    if !exists {
        return zero, application.ErrHandlerNotFound
    }

    handler, ok := rawHandler.(CommandHandler[TCommand, TResult])
    if !ok {
        return zero, application.ErrHandlerNotFound
    }

    finalHandler := func(ctx context.Context, c Command) (any, error) {
        return handler.Handle(ctx, c.(TCommand))
    }

    // apply middleware
    for i := len(bus.middleware) - 1; i >= 0; i-- {
        finalHandler = bus.middleware[i](finalHandler)
    }

    result, err := finalHandler(ctx, cmd)
    if err != nil {
        return zero, err
    }

    typedResult, ok := result.(TResult)
    if !ok {
        return zero, application.ErrInvalidResult
    }

    return typedResult, nil
}

func (bus *CommandBus) Use(mw Middleware) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.middleware = append(bus.middleware, mw)
}

func getTypeKey[TCommand Command]() string {
	var t TCommand
	return fmt.Sprintf("%T", t)
}*/
