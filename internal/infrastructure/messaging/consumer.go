package messaging

import (
	"context"

	appmsg "github.com/victorotene80/medilog-api/internal/application/messaging"
)

type HandlerFunc func(ctx context.Context, envelope appmsg.Envelope) error

type MessageConsumer interface {
	Subscribe(messageName string, handler HandlerFunc)
	Consume(ctx context.Context) error
	Close() error
}