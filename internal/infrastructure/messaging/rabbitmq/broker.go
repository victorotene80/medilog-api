package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	appmsg "github.com/victorotene80/medilog-api/internal/application/messaging"
	broker "github.com/victorotene80/medilog-api/internal/infrastructure/messaging"
)

var _ broker.MessageBroker = (*Broker)(nil)

type BrokerConfig struct {
	DSN            string
	Exchange       string
	RetryExchange  string
	DLExchange     string
	PublishTimeout time.Duration
}

type Broker struct {
	cfg     BrokerConfig
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewBroker(cfg BrokerConfig) (*Broker, error) {
	if cfg.Exchange == "" {
		cfg.Exchange = "auth.tasks"
	}
	if cfg.RetryExchange == "" {
		cfg.RetryExchange = "auth.tasks.retry"
	}
	if cfg.DLExchange == "" {
		cfg.DLExchange = "auth.tasks.dlx"
	}
	if cfg.PublishTimeout == 0 {
		cfg.PublishTimeout = 5 * time.Second
	}

	conn, err := amqp.Dial(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq broker: dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq broker: channel: %w", err)
	}

	for _, ex := range []string{cfg.Exchange, cfg.RetryExchange, cfg.DLExchange} {
		if err := ch.ExchangeDeclare(ex, "topic", true, false, false, false, nil); err != nil {
			_ = ch.Close()
			_ = conn.Close()
			return nil, fmt.Errorf("rabbitmq broker: declare exchange %q: %w", ex, err)
		}
	}

	return &Broker{
		cfg:     cfg,
		conn:    conn,
		channel: ch,
	}, nil
}

func (b *Broker) Publish(ctx context.Context, envelope appmsg.Envelope) error {
	payload, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("rabbitmq broker: marshal envelope: %w", err)
	}

	pubCtx, cancel := context.WithTimeout(ctx, b.cfg.PublishTimeout)
	defer cancel()

	err = b.channel.PublishWithContext(
		pubCtx,
		b.cfg.Exchange,
		envelope.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    envelope.ID,
			Timestamp:    envelope.OccurredAt,
			Body:         payload,
			Headers: amqp.Table{
				"x-retry-count": int32(0),
				"message_name":  envelope.Name,
				"message_kind":  string(envelope.Kind),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("rabbitmq broker: publish %q: %w", envelope.Name, err)
	}

	return nil
}

func (b *Broker) Close() error {
	_ = b.channel.Close()
	return b.conn.Close()
}