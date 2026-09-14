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
	EventBroker MessageBroker
	TaskBroker  MessageBroker
}

type Relay struct {
	repo    outboxInfra.OutboxRepository
	brokers Brokers
	cfg     Config
	logger  *zap.Logger
	done    chan struct{}
}

// Done is closed once Run has returned.
func (r *Relay) Done() <-chan struct{} { return r.done }

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
		done:    make(chan struct{}),
	}
}

// Run polls the outbox until ctx is cancelled. It closes the channel returned by
// Done when it has exited, so a caller can wait for an in-flight tick to finish
// before closing the brokers underneath it.
func (r *Relay) Run(ctx context.Context) {
	defer close(r.done)

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

	if len(envelopes) == 0 {
		return nil
	}

	// Logged only when there is work. At the default one-second poll this line
	// was 86,400 entries per day per instance, almost all of them count=0, into
	// the Loki pipeline this project ships. The scheduler's tick already models
	// the right behaviour by logging only when it created something.
	r.logger.Info("relay fetched pending messages",
		zap.Int("count", len(envelopes)),
	)

	for _, env := range envelopes {
		r.logger.Debug("relay processing message",
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
			r.markFailed(ctx, env.ID, err)
			continue
		}

		brokerName := "unknown"
		switch env.Kind {
		case appmsg.KindIntegrationEvent:
			brokerName = "event"
		case appmsg.KindTask:
			brokerName = "task"
		}

		r.logger.Debug("relay publishing message",
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
			r.markFailed(ctx, env.ID, err)
			continue
		}

		r.logger.Debug("relay published message successfully",
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

		r.logger.Debug("relay marked message sent",
			zap.String("id", env.ID),
			zap.String("name", env.Name),
		)
	}

	return nil
}

// resolveBroker picks the broker for an envelope: a per-message override from
// the route table if one is configured, otherwise the default for its kind.
//
// KindIntegrationEvent previously returned the *task* broker — harmless only
// because bootstrap points both brokers at the same exchange, and actively
// misleading because tick() logs "event" for the same message. The route tables
// were threaded from config through bootstrap into this struct and then never
// consulted.
// markFailed records the failure and logs when even that fails. Discarding this
// error meant a row could stay claimed as in_progress with no trace, waiting on
// the reclaim sweep two minutes into the next boot.
func (r *Relay) markFailed(ctx context.Context, id string, cause error) {
	if err := r.repo.MarkFailed(ctx, id, cause); err != nil {
		r.logger.Error("relay: mark failed did not persist",
			zap.String("id", id),
			zap.NamedError("cause", cause),
			zap.Error(err),
		)
	}
}

func (r *Relay) resolveBroker(env appmsg.Envelope) (MessageBroker, error) {
	switch env.Kind {
	case appmsg.KindIntegrationEvent:
		if name, ok := r.cfg.EventRoutes[env.Name]; ok && name != "" {
			return r.namedBroker(name)
		}
		return r.namedBroker(r.cfg.DefaultEventBroker)
	case appmsg.KindTask:
		if name, ok := r.cfg.TaskRoutes[env.Name]; ok && name != "" {
			return r.namedBroker(name)
		}
		return r.namedBroker(r.cfg.DefaultTaskBroker)
	default:
		return nil, fmt.Errorf("unknown message kind: %s", env.Kind)
	}
}

func (r *Relay) namedBroker(name string) (MessageBroker, error) {
	switch name {
	case "event":
		if r.brokers.EventBroker == nil {
			return nil, fmt.Errorf("event broker is nil")
		}
		return r.brokers.EventBroker, nil
	case "task":
		if r.brokers.TaskBroker == nil {
			return nil, fmt.Errorf("task broker is nil")
		}
		return r.brokers.TaskBroker, nil
	default:
		return nil, fmt.Errorf("unknown broker route: %s", name)
	}
}
