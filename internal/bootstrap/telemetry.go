package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"

	"github.com/victorotene80/medilog-api/internal/shared/config"
)

type Telemetry struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	LoggerProvider *sdklog.LoggerProvider
	shutdown       func(context.Context) error
}

func (t *Telemetry) Shutdown(ctx context.Context) {
	if t == nil || t.shutdown == nil {
		return
	}
	if err := t.shutdown(ctx); err != nil {
		_ = err
	}
}

func initializeTelemetry(cfg *config.Config, logger *zap.Logger) (*Telemetry, error) {
	if !cfg.Telemetry.Enabled {
		logger.Info("telemetry disabled — skipping OTEL SDK init")
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", cfg.Telemetry.ServiceName),
			attribute.String("service.version", cfg.Telemetry.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: build resource: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, traceExporterOptions(cfg.Telemetry)...)
	if err != nil {
		return nil, fmt.Errorf("telemetry: trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	metricExporter, err := otlpmetricgrpc.New(ctx, metricExporterOptions(cfg.Telemetry)...)
	if err != nil {
		_ = tp.Shutdown(ctx)
		return nil, fmt.Errorf("telemetry: metric exporter: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(15*time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	logExporter, err := otlploggrpc.New(ctx, logExporterOptions(cfg.Telemetry)...)
	if err != nil {
		_ = tp.Shutdown(ctx)
		_ = mp.Shutdown(ctx)
		return nil, fmt.Errorf("telemetry: log exporter: %w", err)
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(res),
	)
	global.SetLoggerProvider(lp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("telemetry enabled (OTLP gRPC)",
		zap.String("endpoint", cfg.Telemetry.ExporterEndpoint),
		zap.Bool("insecure", cfg.Telemetry.ExporterInsecure),
	)

	shutdown := func(ctx context.Context) error {
		var errs []error
		if err := tp.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		if err := mp.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		if err := lp.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		return errors.Join(errs...)
	}

	return &Telemetry{
		TracerProvider: tp,
		MeterProvider:  mp,
		LoggerProvider: lp,
		shutdown:       shutdown,
	}, nil
}

func traceExporterOptions(cfg config.TelemetryConfig) []otlptracegrpc.Option {
	opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.ExporterEndpoint)}
	if cfg.ExporterInsecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	return opts
}

func metricExporterOptions(cfg config.TelemetryConfig) []otlpmetricgrpc.Option {
	opts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(cfg.ExporterEndpoint)}
	if cfg.ExporterInsecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}
	return opts
}

func logExporterOptions(cfg config.TelemetryConfig) []otlploggrpc.Option {
	opts := []otlploggrpc.Option{otlploggrpc.WithEndpoint(cfg.ExporterEndpoint)}
	if cfg.ExporterInsecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}
	return opts
}
