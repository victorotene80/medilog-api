package messaging

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/victorotene80/medilog-api/internal/application"
)

// handlerEntry stores a command handler behind a Command-typed adapter so the
// concrete TResult is captured at registration time. Lookup is keyed purely by
// the command type, so a TResult mismatch is reported as a clear error at
// execution time instead of being silently misrouted.
type handlerEntry struct {
	commandType reflect.Type
	handle      func(ctx context.Context, cmd Command) (any, error)
}

type CommandBus struct {
	mu         sync.RWMutex
	handlers   map[string]handlerEntry
	middleware []Middleware
}

func NewCommandBus() *CommandBus {
	return &CommandBus{
		handlers:   make(map[string]handlerEntry),
		middleware: []Middleware{},
	}
}

func Register[TCommand Command, TResult any](
	bus *CommandBus,
	handler CommandHandler[TCommand, TResult],
) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	key := canonicalKey[TCommand]()

	if _, exists := bus.handlers[key]; exists {
		return application.ErrHandlerExists
	}

	bus.handlers[key] = handlerEntry{
		commandType: reflect.TypeOf((*TCommand)(nil)).Elem(),
		handle: func(ctx context.Context, cmd Command) (any, error) {
			typed, ok := cmd.(TCommand)
			if !ok {
				return nil, fmt.Errorf(
					"%w: command type mismatch: got %T, handler expects %T",
					application.ErrHandlerNotFound,
					cmd,
					*new(TCommand),
				)
			}
			return handler.Handle(ctx, typed)
		},
	}
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

	key := canonicalKey[TCommand]()

	bus.mu.RLock()
	entry, exists := bus.handlers[key]
	bus.mu.RUnlock()

	if !exists {
		return zero, fmt.Errorf("%w: %s", application.ErrHandlerNotFound, key)
	}

	if entry.commandType.Kind() == reflect.Ptr {
		if reflect.TypeOf(cmd).Kind() != reflect.Ptr {
			return zero, fmt.Errorf("%w: handler for %s expects %s", application.ErrHandlerNotFound, key, entry.commandType)
		}
	}

	finalHandler := entry.handle

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
		return zero, fmt.Errorf(
			"%w: handler for %s returned %T, expected %T",
			application.ErrInvalidResult,
			key,
			result,
			zero,
		)
	}

	return typedResult, nil
}

func (bus *CommandBus) Use(mw Middleware) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.middleware = append(bus.middleware, mw)
}

// canonicalKey returns a stable key for a command type. Pointers and values are
// treated as the same key so `Execute(&Foo{})` matches a handler registered for
// `Foo{}` and vice versa.
func canonicalKey[TCommand Command]() string {
	return typeKey(reflect.TypeOf((*TCommand)(nil)).Elem())
}

func typeKey(t reflect.Type) string {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t == nil {
		return "<nil>"
	}
	pkg := t.PkgPath()
	if pkg == "" {
		pkg = "(anonymous)"
	}
	return pkg + "." + t.Name()
}
