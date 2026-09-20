package bootstrap

import (
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/infrastructure/scheduler"
	"github.com/victorotene80/medilog-api/internal/shared/config"
	"go.uber.org/zap"
)

// initializeScheduler builds the reminder scanner. It no longer starts anything:
// the scan is driven by Cloud Scheduler calling
// POST /internal/scheduler/reminders/tick, because a Cloud Run container gets
// almost no CPU between requests and none at all once it scales to zero, so an
// in-process ticker silently stopped generating reminders.
//
// Returns nil when disabled, which leaves the tick route unregistered rather
// than serving an endpoint that would do nothing.
//
// It still cannot live in initializeMessaging alongside the outbox relay: that
// runs before initializeCommands, and the scanner needs the command bus.
func initializeScheduler(
	bus *messaging.CommandBus,
	cfg *config.Config,
	logger *zap.Logger,
) *scheduler.ReminderScheduler {
	if !cfg.Scheduler.Enabled {
		logger.Info("reminder scheduler disabled — no reminders will be generated")
		return nil
	}

	if cfg.Scheduler.TickToken == "" {
		// Fatal rather than a warning: the deployment asked for reminders, and
		// silently not registering the only route that can produce them is the
		// exact failure mode this whole change exists to remove.
		logger.Fatal("SCHEDULER_ENABLED is true but SCHEDULER_TICK_TOKEN is unset — " +
			"the tick endpoint would be unreachable and no reminders would be generated")
	}

	logger.Info("reminder scanner ready — driven by POST /internal/scheduler/reminders/tick",
		zap.Duration("catchup_window", cfg.Scheduler.CatchupWindow),
		zap.Duration("lead_time", cfg.Scheduler.LeadTime),
		zap.Int("batch_size", cfg.Scheduler.BatchSize),
	)

	return scheduler.New(bus, scheduler.Config{
		Enabled:       cfg.Scheduler.Enabled,
		PollInterval:  cfg.Scheduler.PollInterval,
		CatchupWindow: cfg.Scheduler.CatchupWindow,
		LeadTime:      cfg.Scheduler.LeadTime,
		BatchSize:     cfg.Scheduler.BatchSize,
	}, logger)
}
