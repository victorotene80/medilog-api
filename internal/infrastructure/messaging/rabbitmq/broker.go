package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
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

	// returned records the message ids the broker handed back as unroutable.
	// RabbitMQ sends basic.return before the basic.ack for the same message, so
	// a publisher that waits for its confirmation and then checks this map can
	// tell "the broker accepted and routed it" from "the broker accepted it and
	// threw it away".
	mu       sync.Mutex
	returned map[string]struct{}
	done     chan struct{}
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

	// Without confirm mode PublishWithContext returns nil as soon as the frame
	// reaches the socket, so a broker that dies before persisting the message
	// still looks like a success and the relay marks the outbox row sent.
	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq broker: enable publisher confirms: %w", err)
	}

	b := &Broker{
		cfg:      cfg,
		conn:     conn,
		channel:  ch,
		returned: make(map[string]struct{}),
		done:     make(chan struct{}),
	}

	returns := ch.NotifyReturn(make(chan amqp.Return, 64))
	go b.drainReturns(returns)

	return b, nil
}

func (b *Broker) drainReturns(returns <-chan amqp.Return) {
	for {
		select {
		case <-b.done:
			return
		case ret, ok := <-returns:
			if !ok {
				return
			}
			b.mu.Lock()
			b.returned[ret.MessageId] = struct{}{}
			b.mu.Unlock()
		}
	}
}

// wasReturned reports and clears whether the broker handed this message back as
// unroutable.
func (b *Broker) wasReturned(messageID string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.returned[messageID]; ok {
		delete(b.returned, messageID)
		return true
	}
	return false
}

func (b *Broker) Publish(ctx context.Context, envelope appmsg.Envelope) error {
	payload, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("rabbitmq broker: marshal envelope: %w", err)
	}

	pubCtx, cancel := context.WithTimeout(ctx, b.cfg.PublishTimeout)
	defer cancel()

	// mandatory=true so an envelope whose routing key matches no binding comes
	// back as a return instead of being silently discarded by the exchange.
	confirm, err := b.channel.PublishWithDeferredConfirmWithContext(
		pubCtx,
		b.cfg.Exchange,
		envelope.Name,
		true,
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

	acked, err := confirm.WaitContext(pubCtx)
	if err != nil {
		return fmt.Errorf("rabbitmq broker: await confirm %q: %w", envelope.Name, err)
	}

	if !acked {
		return fmt.Errorf("rabbitmq broker: publish %q was nacked by the broker", envelope.Name)
	}

	// The ack only says the broker took the message. If it was also returned it
	// matched no binding and was dropped, which must not be reported as a
	// successful delivery — the relay marks the outbox row sent on nil.
	if b.wasReturned(envelope.ID) {
		return fmt.Errorf(
			"rabbitmq broker: %q unroutable on exchange %q (no queue bound to this routing key)",
			envelope.Name, b.cfg.Exchange,
		)
	}

	return nil
}

func (b *Broker) Close() error {
	close(b.done)
	_ = b.channel.Close()
	return b.conn.Close()
}
