package logging

import (
	"context"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	otellog "go.opentelemetry.io/otel/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey struct{}

type LoggerProvider struct {
	logger *zap.Logger
}

// NewLoggerProvider builds a zap logger.
// If an OTEL LoggerProvider is supplied, log records are also forwarded to it
// (and on to Loki via the collector) in addition to stdout.
func NewLoggerProvider(otelLP otellog.LoggerProvider) (*LoggerProvider, error) {
	// Base production encoder — writes to stdout as before.
	prodCfg := zap.NewProductionConfig()
	baseCore, err := prodCfg.Build()
	if err != nil {
		return nil, err
	}

	core := baseCore.Core()

	if otelLP != nil {
		// otelzap.NewCore sends every zap record to the OTEL log pipeline.
		otelCore := otelzap.NewCore("medilog-api", otelzap.WithLoggerProvider(otelLP))
		core = zapcore.NewTee(core, otelCore)
	}

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return &LoggerProvider{logger: logger}, nil
}

func (p *LoggerProvider) Logger() *zap.Logger { return p.logger }
func (p *LoggerProvider) Sync()               { _ = p.logger.Sync() }

func WithContext(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

func FromContext(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok && l != nil {
		return l
	}
	return zap.NewNop()
}
