package messaging

import (
	"context"

	appmsg "github.com/victorotene80/medilog-api/internal/application/messaging"
)

type MessageBroker interface {
	Publish(ctx context.Context, envelope appmsg.Envelope) error
	Close() error
}