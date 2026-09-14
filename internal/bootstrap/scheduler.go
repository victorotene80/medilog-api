package bootstrap

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/infrastructure/scheduler"
	"github.com/victorotene80/medilog-api/internal/shared/config"
	"go.uber.org/zap"
)

// initializeScheduler starts the reminder ticker and returns its stop function.
//
// It cannot live in initializeMessaging alongside the outbox relay: that runs
// before initializeCommands, and the scheduler needs the command bus. It is
// therefore started from app.go once the bus exists, and must be stopped
// *before* messaging so an in-flight tick can finish publishing.
func initializeScheduler(
	bus *messaging.CommandBus,
	cfg *config.Config,
	logger *zap.Logger,
) func() {
	if !cfg.Scheduler.Enabled {
		logger.Info("reminder scheduler disabled — no reminders will be generated")
		return func() {}
	}

	sched := scheduler.New(bus, scheduler.Config{
		Enabled:       cfg.Scheduler.Enabled,
		PollInterval:  cfg.Scheduler.PollInterval,
		CatchupWindow: cfg.Scheduler.CatchupWindow,
		LeadTime:      cfg.Scheduler.LeadTime,
		BatchSize:     cfg.Scheduler.BatchSize,
	}, logger)

	ctx, cancel := context.WithCancel(context.Background())
	go sched.Run(ctx)

	return cancel
}
