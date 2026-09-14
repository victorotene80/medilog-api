package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	appmsg "github.com/victorotene80/medilog-api/internal/application/messaging"
	consumer "github.com/victorotene80/medilog-api/internal/infrastructure/messaging"
)

var _ consumer.MessageConsumer = (*Consumer)(nil)

type ConsumerConfig struct {
	DSN           string
	Exchange      string
	RetryExchange string
	DLExchange    string
	MaxRetries    int
	RetryDelay    time.Duration
	Prefetch      int
}

type subscription struct {
	name    string
	handler consumer.HandlerFunc
}

type Consumer struct {
	cfg           ConsumerConfig
	mu            sync.Mutex
	subscriptions []subscription

	conn    *amqp.Connection
	channel *amqp.Channel

	cancel context.CancelFunc
	wg     sync.WaitGroup
	once   sync.Once
}

func NewConsumer(cfg ConsumerConfig) (*Consumer, error) {
	if cfg.Exchange == "" {
		cfg.Exchange = "auth.tasks"
	}
	if cfg.RetryExchange == "" {
		cfg.RetryExchange = "auth.tasks.retry"
	}
	if cfg.DLExchange == "" {
		cfg.DLExchange = "auth.tasks.dlx"
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 15 * time.Second
	}
	if cfg.Prefetch == 0 {
		cfg.Prefetch = 10
	}

	conn, err := amqp.Dial(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq consumer: dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq consumer: channel: %w", err)
	}

	if err := ch.Qos(cfg.Prefetch, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq consumer: qos: %w", err)
	}

	for _, ex := range []string{cfg.Exchange, cfg.RetryExchange, cfg.DLExchange} {
		if err := ch.ExchangeDeclare(ex, "topic", true, false, false, false, nil); err != nil {
			_ = ch.Close()
			_ = conn.Close()
			return nil, fmt.Errorf("rabbitmq consumer: exchange declare %q: %w", ex, err)
		}
	}

	return &Consumer{
		cfg:     cfg,
		conn:    conn,
		channel: ch,
	}, nil
}

func (c *Consumer) Subscribe(messageName string, handler consumer.HandlerFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	mainQueue := queueName(messageName)
	retryQueue := retryQueueName(messageName)
	dlq := dlqName(messageName)

	_, _ = c.channel.QueueDeclare(mainQueue, true, false, false, false, nil)
	_ = c.channel.QueueBind(mainQueue, messageName, c.cfg.Exchange, false, nil)

	_, _ = c.channel.QueueDeclare(retryQueue, true, false, false, false, amqp.Table{
		"x-message-ttl":             int32(c.cfg.RetryDelay / time.Millisecond),
		"x-dead-letter-exchange":    c.cfg.Exchange,
		"x-dead-letter-routing-key": messageName,
	})
	_ = c.channel.QueueBind(retryQueue, messageName, c.cfg.RetryExchange, false, nil)

	_, _ = c.channel.QueueDeclare(dlq, true, false, false, false, nil)
	_ = c.channel.QueueBind(dlq, messageName, c.cfg.DLExchange, false, nil)
	c.subscriptions = append(c.subscriptions, subscription{
		name:    messageName,
		handler: handler,
	})
}

func (c *Consumer) Consume(ctx context.Context) error {
	c.mu.Lock()
	subs := make([]subscription, len(c.subscriptions))
	copy(subs, c.subscriptions)
	c.mu.Unlock()

	if len(subs) == 0 {
		return fmt.Errorf("rabbitmq consumer: no subscriptions")
	}

	runCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	errCh := make(chan error, len(subs))

	for _, sub := range subs {
		c.wg.Add(1)
		go func(s subscription) {
			defer c.wg.Done()
			if err := c.consumeQueue(runCtx, s); err != nil {
				errCh <- err
			}
		}(sub)
	}

	go func() {
		c.wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if err != nil {
			cancel()
			c.wg.Wait()
			return err
		}
	}

	return nil
}

// ErrDeliveriesClosed reports that the broker closed the delivery channel while
// the consumer was still meant to be running.
var ErrDeliveriesClosed = errors.New("rabbitmq consumer: broker closed the delivery channel")

func (c *Consumer) consumeQueue(ctx context.Context, sub subscription) error {
	queue := queueName(sub.name)

	deliveries, err := c.channel.Consume(
		queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("rabbitmq consumer: consume queue %q: %w", queue, err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-deliveries:
			if !ok {
				// amqp091 closes this channel whenever the underlying channel or
				// connection closes — a broker restart, a heartbeat miss, a
				// network blip. Returning nil here made that indistinguishable
				// from the ctx.Done() shutdown above, so every consumer goroutine
				// exited, Consume returned nil, and the caller logged nothing:
				// the API stayed up, healthy and silent, consuming nothing until
				// someone redeployed it.
				//
				// NOTE: this makes the failure visible, it does not recover from
				// it. NewConsumer dials once and there is no reconnect loop, so
				// the process still needs restarting to resume consumption.
				return fmt.Errorf(
					"%w: queue %q", ErrDeliveriesClosed, queue,
				)
			}
			if err := c.handle(ctx, sub.name, d, sub.handler); err != nil {
				return err
			}
		}
	}
}

func (c *Consumer) handle(
	ctx context.Context,
	messageName string,
	d amqp.Delivery,
	handler consumer.HandlerFunc,
) error {
	var env appmsg.Envelope
	if err := json.Unmarshal(d.Body, &env); err != nil {
		_ = d.Nack(false, false)
		return nil
	}

	if err := handler(ctx, env); err != nil {
		retryCount := retryCountFromHeaders(d.Headers)

		if retryCount >= c.cfg.MaxRetries {
			if err := c.publishToDLQ(messageName, env, retryCount+1); err != nil {
				return err
			}
			_ = d.Ack(false)
			return nil
		}

		if err := c.publishToRetry(messageName, env, retryCount+1); err != nil {
			return err
		}
		_ = d.Ack(false)
		return nil
	}

	_ = d.Ack(false)
	return nil
}

func (c *Consumer) publishToRetry(messageName string, env appmsg.Envelope, retryCount int) error {
	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("rabbitmq consumer: marshal retry envelope: %w", err)
	}

	err = c.channel.Publish(
		c.cfg.RetryExchange,
		messageName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    env.ID,
			Timestamp:    time.Now().UTC(),
			Body:         body,
			Headers: amqp.Table{
				"x-retry-count": int32(retryCount),
				"message_name":  env.Name,
				"message_kind":  string(env.Kind),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("rabbitmq consumer: publish to retry: %w", err)
	}

	return nil
}

func (c *Consumer) publishToDLQ(messageName string, env appmsg.Envelope, retryCount int) error {
	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("rabbitmq consumer: marshal dlq envelope: %w", err)
	}

	err = c.channel.Publish(
		c.cfg.DLExchange,
		messageName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    env.ID,
			Timestamp:    time.Now().UTC(),
			Body:         body,
			Headers: amqp.Table{
				"x-retry-count": int32(retryCount),
				"message_name":  env.Name,
				"message_kind":  string(env.Kind),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("rabbitmq consumer: publish to dlq: %w", err)
	}

	return nil
}

func retryCountFromHeaders(headers amqp.Table) int {
	v, ok := headers["x-retry-count"]
	if !ok {
		return 0
	}

	switch n := v.(type) {
	case int32:
		return int(n)
	case int64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}

func queueName(messageName string) string {
	return "q." + messageName
}

func retryQueueName(messageName string) string {
	return "q.retry." + messageName
}

func dlqName(messageName string) string {
	return "q.dlq." + messageName
}

func (c *Consumer) Close() error {
	c.once.Do(func() {
		if c.cancel != nil {
			c.cancel()
		}
		c.wg.Wait()
		_ = c.channel.Close()
		_ = c.conn.Close()
	})
	return nil
}
