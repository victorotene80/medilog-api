package messaging

import (
	"context"
	"errors"
	"testing"

	"github.com/victorotene80/medilog-api/internal/application"
)

type testCommand struct{ Value string }
type testCommandResult struct{ Upper string }

func TestExecuteRegisteredCommand(t *testing.T) {
	bus := NewCommandBus()
	MustRegister(bus, CommandHandlerFunc[testCommand, testCommandResult](
		func(_ context.Context, cmd testCommand) (testCommandResult, error) {
			return testCommandResult{Upper: cmd.Value + "!"}, nil
		},
	))

	res, err := Execute[testCommand, testCommandResult](bus, context.Background(), testCommand{Value: "ok"})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if res.Upper != "ok!" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestExecuteResultTypeMismatchIsExplicit(t *testing.T) {
	bus := NewCommandBus()
	MustRegister(bus, CommandHandlerFunc[testCommand, testCommandResult](
		func(_ context.Context, cmd testCommand) (testCommandResult, error) {
			return testCommandResult{Upper: cmd.Value}, nil
		},
	))

	_, err := Execute[testCommand, string](bus, context.Background(), testCommand{Value: "x"})
	if !errors.Is(err, application.ErrInvalidResult) {
		t.Fatalf("expected ErrInvalidResult, got %v", err)
	}
}

func TestExecuteUnregisteredCommand(t *testing.T) {
	bus := NewCommandBus()
	_, err := Execute[testCommand, testCommandResult](bus, context.Background(), testCommand{})
	if !errors.Is(err, application.ErrHandlerNotFound) {
		t.Fatalf("expected ErrHandlerNotFound, got %v", err)
	}
}

func TestRegisterAndExecutePointerCommand(t *testing.T) {
	bus := NewCommandBus()
	MustRegister(bus, CommandHandlerFunc[*testCommand, testCommandResult](
		func(_ context.Context, cmd *testCommand) (testCommandResult, error) {
			return testCommandResult{Upper: cmd.Value + "*"}, nil
		},
	))

	res, err := Execute[*testCommand, testCommandResult](bus, context.Background(), &testCommand{Value: "p"})
	if err != nil {
		t.Fatalf("execute by pointer: %v", err)
	}
	if res.Upper != "p*" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestPointerAndValueCommandsShareOneKey(t *testing.T) {
	bus := NewCommandBus()
	MustRegister(bus, CommandHandlerFunc[testCommand, testCommandResult](
		func(_ context.Context, cmd testCommand) (testCommandResult, error) {
			return testCommandResult{Upper: cmd.Value}, nil
		},
	))

	// Registering the pointer variant of the same command must collide with the
	// value registration (canonicalized key), not silently create a second entry.
	if err := Register[*testCommand, testCommandResult](bus, CommandHandlerFunc[*testCommand, testCommandResult](
		func(_ context.Context, cmd *testCommand) (testCommandResult, error) {
			return testCommandResult{Upper: cmd.Value}, nil
		},
	)); !errors.Is(err, application.ErrHandlerExists) {
		t.Fatalf("expected ErrHandlerExists, got %v", err)
	}
}

func TestExecuteValueCommandAgainstPointerHandlerMismatch(t *testing.T) {
	bus := NewCommandBus()
	MustRegister(bus, CommandHandlerFunc[testCommand, testCommandResult](
		func(_ context.Context, cmd testCommand) (testCommandResult, error) {
			return testCommandResult{Upper: cmd.Value}, nil
		},
	))

	// A pointer execution against a value-registered handler must error, not panic.
	_, err := Execute[*testCommand, testCommandResult](bus, context.Background(), &testCommand{Value: "x"})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

type CommandHandlerFunc[TCommand Command, TResult any] func(ctx context.Context, cmd TCommand) (TResult, error)

func (f CommandHandlerFunc[TCommand, TResult]) Handle(ctx context.Context, cmd TCommand) (TResult, error) {
	return f(ctx, cmd)
}
