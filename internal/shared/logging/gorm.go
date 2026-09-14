package logging

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

type GormConfig struct {
	SlowThreshold             time.Duration
	LogLevel                  gormlog.LogLevel
	IgnoreRecordNotFoundError bool
}

type GormLogger struct {
	logger *zap.Logger
	config GormConfig
}

func NewGormLogger(logger *zap.Logger, config GormConfig) *GormLogger {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config.SlowThreshold == 0 {
		config.SlowThreshold = 200 * time.Millisecond
	}

	if config.LogLevel == 0 {
		config.LogLevel = gormlog.Warn
	}

	return &GormLogger{
		logger: logger.Named("gorm"),
		config: config,
	}
}

func (l *GormLogger) LogMode(level gormlog.LogLevel) gormlog.Interface {
	next := *l
	next.config.LogLevel = level
	return &next
}

func (l *GormLogger) Info(_ context.Context, msg string, args ...interface{}) {
	if l.config.LogLevel < gormlog.Info {
		return
	}

	l.logger.Info("gorm info", zap.String("message", fmt.Sprintf(msg, args...)))
}

func (l *GormLogger) Warn(_ context.Context, msg string, args ...interface{}) {
	if l.config.LogLevel < gormlog.Warn {
		return
	}

	l.logger.Warn("gorm warning", zap.String("message", fmt.Sprintf(msg, args...)))
}

func (l *GormLogger) Error(_ context.Context, msg string, args ...interface{}) {
	if l.config.LogLevel < gormlog.Error {
		return
	}

	l.logger.Error("gorm error", zap.String("message", fmt.Sprintf(msg, args...)))
}

func (l *GormLogger) Trace(
	_ context.Context,
	begin time.Time,
	fc func() (sql string, rowsAffected int64),
	err error,
) {
	if l.config.LogLevel <= gormlog.Silent {
		return
	}

	elapsed := time.Since(begin)

	if err != nil && l.config.LogLevel >= gormlog.Error {
		if l.config.IgnoreRecordNotFoundError && errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}

		sql, rowsAffected := fc()
		fields := []zap.Field{
			zap.Error(err),
			zap.String("error_type", fmt.Sprintf("%T", err)),
			zap.String("sql", sql),
			zap.Int64("rows_affected", rowsAffected),
			zap.Duration("elapsed", elapsed),
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// pg_detail is deliberately omitted: Postgres puts the offending
			// values in it ("Key (email)=(alice@example.com) already exists."),
			// which is the same user data ParamsFilter strips from the SQL above.
			// pg_constraint and pg_column identify the failure without it.
			fields = append(fields,
				zap.String("pg_code", pgErr.Code),
				zap.String("pg_message", pgErr.Message),
				zap.String("pg_table", pgErr.TableName),
				zap.String("pg_column", pgErr.ColumnName),
				zap.String("pg_constraint", pgErr.ConstraintName),
			)
		}

		l.logger.Error("gorm query failed", fields...)
		return
	}

	if l.config.SlowThreshold > 0 && elapsed > l.config.SlowThreshold && l.config.LogLevel >= gormlog.Warn {
		sql, rowsAffected := fc()
		l.logger.Warn("gorm slow query",
			zap.Duration("threshold", l.config.SlowThreshold),
			zap.String("sql", sql),
			zap.Int64("rows_affected", rowsAffected),
			zap.Duration("elapsed", elapsed),
		)
		return
	}

	if l.config.LogLevel == gormlog.Info {
		sql, rowsAffected := fc()
		l.logger.Info("gorm query",
			zap.String("sql", sql),
			zap.Int64("rows_affected", rowsAffected),
			zap.Duration("elapsed", elapsed),
		)
	}
}

func (l *GormLogger) ParamsFilter(_ context.Context, sql string, _ ...interface{}) (string, []interface{}) {
	return sql, nil
}
