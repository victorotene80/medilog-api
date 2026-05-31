package bootstrap

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/victorotene80/medilog-api/internal/shared/config"
)

// Telemetry holds the three SDK providers and their shared gRPC connection.
// Call Shutdown when the process exits to flush any buffered data.
type Telemetry struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	LoggerProvider *sdklog.LoggerProvider
	shutdown       func(context.Context) error
}

// Shutdown flushes and stops all providers. Safe to call on a nil receiver
// (no-op), so callers don't need to guard against telemetry being disabled.
func (t *Telemetry) Shutdown(ctx context.Context) {
	if t == nil {
		return
	}
	if err := t.shutdown(ctx); err != nil {
		// Best-effort — we're already shutting down.
		_ = err
	}
}

// initializeTelemetry wires the OTEL SDK when cfg.Telemetry.Enabled is true.
// When disabled it returns nil, and callers treat that as "no-op telemetry".
func initializeTelemetry(cfg *config.Config, logger *zap.Logger) (*Telemetry, error) {
	if !cfg.Telemetry.Enabled {
		logger.Info("telemetry disabled — skipping OTEL SDK init")
		return nil, nil
	}

	ctx := context.Background()

	dialOpts := []grpc.DialOption{grpc.WithBlock()}
	if cfg.Telemetry.ExporterInsecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	connCtx, connCancel := context.WithTimeout(ctx, 5*time.Second)
	defer connCancel()

	conn, err := grpc.DialContext(connCtx, cfg.Telemetry.ExporterEndpoint, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("otel: dial collector at %s: %w", cfg.Telemetry.ExporterEndpoint, err)
	}

	res, err := sdkresource.New(ctx,
		sdkresource.WithOS(),
		sdkresource.WithProcess(),
		sdkresource.WithAttributes(
			semconv.ServiceName(cfg.Telemetry.ServiceName),
			semconv.ServiceVersion(cfg.Telemetry.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otel: build resource: %w", err)
	}

	traceExp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("otel: trace exporter: %w", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	metricExp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("otel: metric exporter: %w", err)
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp,
			sdkmetric.WithInterval(15*time.Second),
		)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	logExp, err := otlploggrpc.New(ctx, otlploggrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("otel: log exporter: %w", err)
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		sdklog.WithResource(res),
	)
	global.SetLoggerProvider(lp)

	logger.Info("telemetry initialised",
		zap.String("endpoint", cfg.Telemetry.ExporterEndpoint),
		zap.String("service", cfg.Telemetry.ServiceName),
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
		if err := conn.Close(); err != nil {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			return fmt.Errorf("otel shutdown errors: %v", errs)
		}
		return nil
	}

	return &Telemetry{
		TracerProvider: tp,
		MeterProvider:  mp,
		LoggerProvider: lp,
		shutdown:       shutdown,
	}, nil
}
