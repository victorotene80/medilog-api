package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	kafkago "github.com/segmentio/kafka-go"

	appmsg "github.com/victorotene80/medilog-api/internal/application/messaging"
	consumer "github.com/victorotene80/medilog-api/internal/infrastructure/messaging"
)

var _ consumer.MessageConsumer = (*Consumer)(nil)

type ConsumerConfig struct {
	Brokers     []string
	GroupID     string
	TopicPrefix string
}

type Consumer struct {
	cfg      ConsumerConfig
	mu       sync.RWMutex
	handlers map[string]consumer.HandlerFunc

	cancel context.CancelFunc
	wg     sync.WaitGroup
	once   sync.Once

	resolver TopicResolver
}

func NewConsumer(cfg ConsumerConfig) *Consumer {
	return &Consumer{
		cfg:      cfg,
		handlers: make(map[string]consumer.HandlerFunc),
		resolver: TopicResolver{Prefix: cfg.TopicPrefix},
	}
}

func (c *Consumer) Subscribe(messageName string, handler consumer.HandlerFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fullTopic := c.resolver.NormalizeSubscription(messageName)
	c.handlers[fullTopic] = handler
}

func (c *Consumer) Consume(ctx context.Context) error {
	c.mu.RLock()
	topics := make(map[string]consumer.HandlerFunc, len(c.handlers))
	for topic, handler := range c.handlers {
		topics[topic] = handler
	}
	c.mu.RUnlock()

	if len(topics) == 0 {
		return fmt.Errorf("kafka consumer: no topics subscribed")
	}

	runCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	errCh := make(chan error, len(topics))

	for topic, handler := range topics {
		c.wg.Add(1)
		go func(t string, h consumer.HandlerFunc) {
			defer c.wg.Done()
			if err := c.readTopic(runCtx, t, h); err != nil {
				errCh <- err
			}
		}(topic, handler)
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

func (c *Consumer) readTopic(ctx context.Context, topic string, handler consumer.HandlerFunc) error {
	r := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: c.cfg.Brokers,
		GroupID: c.cfg.GroupID,
		Topic:   topic,
	})
	defer r.Close()

	for {
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("kafka consumer: fetch from %q: %w", topic, err)
		}

		var env appmsg.Envelope
		if err := json.Unmarshal(msg.Value, &env); err != nil {
			_ = r.CommitMessages(ctx, msg)
			continue
		}

		if err := handler(ctx, env); err != nil {
			continue
		}

		if err := r.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("kafka consumer: commit on %q: %w", topic, err)
		}
	}
}

func (c *Consumer) Close() error {
	c.once.Do(func() {
		if c.cancel != nil {
			c.cancel()
		}
		c.wg.Wait()
	})
	return nil
}