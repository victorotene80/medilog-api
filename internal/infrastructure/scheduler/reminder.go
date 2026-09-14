// Package scheduler holds background tickers. It deliberately contains no
// domain logic: each worker dispatches a command and logs the outcome, so the
// interesting behaviour stays in an application handler that can be unit-tested
// without a database.
package scheduler

import (
	"context"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"go.uber.org/zap"
)

type Config struct {
	Enabled bool
	// PollInterval is 1 minute rather than the relay's 1 second: reminder
	// granularity is HH:MM, so ticking faster multiplies the scan cost without
	// improving anything.
	PollInterval time.Duration
	// CatchupWindow is how far back each tick looks. This is the entire
	// catch-up mechanism — there is no watermark table, because a watermark is
	// a second source of truth that drifts from the dedupe index and goes stale
	// after a restore from backup.
	CatchupWindow time.Duration
	// LeadTime generates slightly ahead of now, so a reminder is already in the
	// inbox when its moment arrives.
	LeadTime  time.Duration
	BatchSize int
}

type ReminderScheduler struct {
	bus    *messaging.CommandBus
	cfg    Config
	logger *zap.Logger
}

func New(bus *messaging.CommandBus, cfg Config, logger *zap.Logger) *ReminderScheduler {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = time.Minute
	}
	if cfg.CatchupWindow <= 0 {
		cfg.CatchupWindow = 6 * time.Hour
	}
	// Clamped so a long outage cannot dump days of stale reminders into an
	// inbox all at once. A three-day-old dose reminder is noise.
	if cfg.CatchupWindow > 24*time.Hour {
		cfg.CatchupWindow = 24 * time.Hour
	}
	if cfg.LeadTime <= 0 {
		cfg.LeadTime = 5 * time.Minute
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 500
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	return &ReminderScheduler{bus: bus, cfg: cfg, logger: logger}
}

func (s *ReminderScheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()

	s.logger.Info("reminder scheduler started",
		zap.Duration("poll_interval", s.cfg.PollInterval),
		zap.Duration("catchup_window", s.cfg.CatchupWindow),
		zap.Duration("lead_time", s.cfg.LeadTime),
		zap.Int("batch_size", s.cfg.BatchSize),
	)

	// Run once immediately so a restart does not leave a poll-interval hole.
	if err := s.tick(ctx); err != nil {
		s.logger.Error("reminder scheduler tick error", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("reminder scheduler stopping")
			return
		case <-ticker.C:
			if err := s.tick(ctx); err != nil {
				s.logger.Error("reminder scheduler tick error", zap.Error(err))
			}
		}
	}
}

func (s *ReminderScheduler) tick(ctx context.Context) error {
	now := time.Now().UTC()

	result, err := messaging.Execute[
		command.GenerateDueRemindersCommand,
		*dto.ReminderRunResultDTO,
	](s.bus, ctx, command.GenerateDueRemindersCommand{
		From:  now.Add(-s.cfg.CatchupWindow),
		To:    now.Add(s.cfg.LeadTime),
		Limit: s.cfg.BatchSize,
	})
	if err != nil {
		return err
	}

	if result == nil || result.LockNotAcquired {
		return nil
	}

	// In steady state almost everything is skipped, so only log when something
	// was actually created.
	if result.MedicationCreated > 0 || result.AppointmentCreated > 0 {
		s.logger.Info("reminders generated",
			zap.Int("scanned", result.Scanned),
			zap.Int("medication_created", result.MedicationCreated),
			zap.Int("appointment_created", result.AppointmentCreated),
			zap.Int("skipped", result.Skipped),
		)
	}

	return nil
}
