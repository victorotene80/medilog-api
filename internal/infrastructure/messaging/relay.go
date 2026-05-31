package messaging

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	appmsg "github.com/victorotene80/medilog-api/internal/application/messaging"
	outboxInfra "github.com/victorotene80/medilog-api/internal/infrastructure/messaging/outbox"
)

type Config struct {
	PollInterval       time.Duration
	BatchSize          int
	ReclaimAfter       time.Duration
	DefaultEventBroker string
	DefaultTaskBroker  string
	EventRoutes        map[string]string
	TaskRoutes         map[string]string
}

type Brokers struct {
	//EventBroker MessageBroker
	TaskBroker MessageBroker
}

type Relay struct {
	repo    outboxInfra.OutboxRepository
	brokers Brokers
	cfg     Config
	logger  *zap.Logger
}

func New(
	repo outboxInfra.OutboxRepository,
	brokers Brokers,
	cfg Config,
	logger *zap.Logger,
) *Relay {
	if cfg.PollInterval == 0 {
		cfg.PollInterval = time.Second
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 50
	}
	if cfg.ReclaimAfter == 0 {
		cfg.ReclaimAfter = 2 * time.Minute
	}
	if cfg.DefaultEventBroker == "" {
		cfg.DefaultEventBroker = "event"
	}
	if cfg.DefaultTaskBroker == "" {
		cfg.DefaultTaskBroker = "task"
	}
	if cfg.EventRoutes == nil {
		cfg.EventRoutes = make(map[string]string)
	}
	if cfg.TaskRoutes == nil {
		cfg.TaskRoutes = make(map[string]string)
	}

	return &Relay{
		repo:    repo,
		brokers: brokers,
		cfg:     cfg,
		logger:  logger,
	}
}

func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.cfg.PollInterval)
	defer ticker.Stop()

	r.logger.Info("outbox relay started",
		zap.Duration("poll_interval", r.cfg.PollInterval),
		zap.Int("batch_size", r.cfg.BatchSize),
		zap.Duration("reclaim_after", r.cfg.ReclaimAfter),
	)

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("outbox relay stopping")
			return
		case <-ticker.C:
			if err := r.tick(ctx); err != nil {
				r.logger.Error("outbox relay tick error", zap.Error(err))
			}
		}
	}
}

func (r *Relay) tick(ctx context.Context) error {
	reclaimed, err := r.repo.ReclaimStaleInProgress(
		ctx,
		time.Now().UTC().Add(-r.cfg.ReclaimAfter),
		r.cfg.BatchSize,
	)
	if err != nil {
		return fmt.Errorf("relay: reclaim stale in-progress: %w", err)
	}
	if reclaimed > 0 {
		r.logger.Warn("relay reclaimed stale messages", zap.Int("count", reclaimed))
	}

	envelopes, err := r.repo.FetchUnprocessed(ctx, r.cfg.BatchSize)
	if err != nil {
		return fmt.Errorf("relay: fetch unprocessed: %w", err)
	}

	r.logger.Info("relay fetched pending messages",
		zap.Int("count", len(envelopes)),
	)

	for _, env := range envelopes {
		r.logger.Info("relay processing message",
			zap.String("id", env.ID),
			zap.String("name", env.Name),
			zap.String("kind", string(env.Kind)),
			zap.Int64("aggregate_id", env.AggregateID),
			zap.String("aggregate_type", env.AggregateType),
		)

		if err := r.repo.MarkInProgress(ctx, env.ID); err != nil {
			r.logger.Warn("relay: mark in-progress failed",
				zap.String("id", env.ID),
				zap.Error(err),
			)
			continue
		}

		broker, err := r.resolveBroker(env)
		if err != nil {
			r.logger.Error("relay: resolve broker failed",
				zap.String("id", env.ID),
				zap.String("name", env.Name),
				zap.String("kind", string(env.Kind)),
				zap.Error(err),
			)
			_ = r.repo.MarkFailed(ctx, env.ID)
			continue
		}

		brokerName := "unknown"
		switch env.Kind {
		case appmsg.KindIntegrationEvent:
			brokerName = "event"
		case appmsg.KindTask:
			brokerName = "task"
		}

		r.logger.Info("relay publishing message",
			zap.String("id", env.ID),
			zap.String("name", env.Name),
			zap.String("kind", string(env.Kind)),
			zap.String("broker", brokerName),
		)

		if err := broker.Publish(ctx, env); err != nil {
			r.logger.Error("relay: publish failed",
				zap.String("id", env.ID),
				zap.String("name", env.Name),
				zap.String("kind", string(env.Kind)),
				zap.String("broker", brokerName),
				zap.Error(err),
			)
			_ = r.repo.MarkFailed(ctx, env.ID)
			continue
		}

		r.logger.Info("relay published message successfully",
			zap.String("id", env.ID),
			zap.String("name", env.Name),
			zap.String("kind", string(env.Kind)),
			zap.String("broker", brokerName),
		)

		if err := r.repo.MarkSent(ctx, env.ID); err != nil {
			r.logger.Error("relay: mark sent failed",
				zap.String("id", env.ID),
				zap.Error(err),
			)
			continue
		}

		r.logger.Info("relay marked message sent",
			zap.String("id", env.ID),
			zap.String("name", env.Name),
		)
	}

	return nil
}

func (r *Relay) resolveBroker(env appmsg.Envelope) (MessageBroker, error) {
	switch env.Kind {
	case appmsg.KindIntegrationEvent:
		route := r.cfg.DefaultEventBroker
		if v, ok := r.cfg.EventRoutes[env.Name]; ok {
			route = v
		}
		return r.namedBroker(route)
	case appmsg.KindTask:
		route := r.cfg.DefaultTaskBroker
		if v, ok := r.cfg.TaskRoutes[env.Name]; ok {
			route = v
		}
		return r.namedBroker(route)
	default:
		return nil, fmt.Errorf("unknown message kind: %s", env.Kind)
	}
}

func (r *Relay) namedBroker(name string) (MessageBroker, error) {
	switch name {
	//case "event":
	///	if r.brokers.EventBroker == nil {
	//		return nil, fmt.Errorf("event broker is nil")
	//	}
	//	return r.brokers.EventBroker, nil
	case "task":
		if r.brokers.TaskBroker == nil {
			return nil, fmt.Errorf("task broker is nil")
		}
		return r.brokers.TaskBroker, nil
	default:
		return nil, fmt.Errorf("unknown broker route: %s", name)
	}
}
