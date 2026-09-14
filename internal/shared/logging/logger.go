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

// NewLoggerProvider builds a zap logger. When otelLP is non-nil it also
// dual-writes log records into the OpenTelemetry pipeline via an otelzap core.
func NewLoggerProvider(otelLP otellog.LoggerProvider) (*LoggerProvider, error) {
	prodCfg := zap.NewProductionConfig()
	baseCore, err := prodCfg.Build()
	if err != nil {
		return nil, err
	}

	core := baseCore.Core()
	if otelLP != nil {
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
