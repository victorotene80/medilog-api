package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	appmsg "github.com/victorotene80/medilog-api/internal/application/messaging"
	broker "github.com/victorotene80/medilog-api/internal/infrastructure/messaging"
)

var _ broker.MessageBroker = (*Broker)(nil)

type BrokerConfig struct {
	Brokers      []string
	TopicPrefix  string
	WriteTimeout time.Duration
}

type Broker struct {
	cfg      BrokerConfig
	writer   *kafkago.Writer
	resolver TopicResolver
}

func NewBroker(cfg BrokerConfig) *Broker {
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 10 * time.Second
	}

	return &Broker{
		cfg: cfg,
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(cfg.Brokers...),
			Balancer:     &kafkago.LeastBytes{},
			WriteTimeout: cfg.WriteTimeout,
		},
		resolver: TopicResolver{Prefix: cfg.TopicPrefix},
	}
}

func (b *Broker) Publish(ctx context.Context, envelope appmsg.Envelope) error {
	payload, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("kafka broker: marshal envelope: %w", err)
	}

	topic := b.resolver.TopicForMessage(envelope.Name)

	err = b.writer.WriteMessages(ctx, kafkago.Message{
		Topic: topic,
		Key:   []byte(fmt.Sprintf("%d", envelope.AggregateID)),
		Value: payload,
		Time:  envelope.OccurredAt,
		Headers: []kafkago.Header{
			{Key: "message_name", Value: []byte(envelope.Name)},
			{Key: "message_kind", Value: []byte(envelope.Kind)},
			{Key: "aggregate_type", Value: []byte(envelope.AggregateType)},
		},
	})
	if err != nil {
		return fmt.Errorf("kafka broker: write to topic %q: %w", topic, err)
	}

	return nil
}

func (b *Broker) Close() error {
	return b.writer.Close()
}