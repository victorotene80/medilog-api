package bootstrap

import (
	"context"

	"go.uber.org/zap"
)

// Telemetry — COMMENTED OUT: OTel SDK stack (collector, tracing, metrics, logs)
// Uncomment the full implementation below to re-enable.
// ================================================================================
// Original full implementation with all OTel imports and OTLP exporters:
// import (
// 	"fmt"
// 	"time"
//
// 	"go.opentelemetry.io/otel"
// 	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
// 	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
// 	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
// 	"go.opentelemetry.io/otel/log/global"
// 	"go.opentelemetry.io/otel/propagation"
// 	sdklog "go.opentelemetry.io/otel/sdk/log"
// 	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
// 	sdkresource "go.opentelemetry.io/otel/sdk/resource"
// 	sdktrace "go.opentelemetry.io/otel/sdk/trace"
// 	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/credentials/insecure"
//
// 	"github.com/victorotene80/medilog-api/internal/shared/config"
// )
//
// type Telemetry struct {
// 	TracerProvider *sdktrace.TracerProvider
// 	MeterProvider  *sdkmetric.MeterProvider
// 	LoggerProvider *sdklog.LoggerProvider
// 	shutdown       func(context.Context) error
// }
//
// func (t *Telemetry) Shutdown(ctx context.Context) {
// 	if t == nil { return }
// 	if err := t.shutdown(ctx); err != nil { _ = err }
// }
//
// func initializeTelemetry(cfg *config.Config, logger *zap.Logger) (*Telemetry, error) {
// 	if !cfg.Telemetry.Enabled {
// 		logger.Info("telemetry disabled — skipping OTEL SDK init")
// 		return nil, nil
// 	}
// 	...
// }
// ================================================================================

type Telemetry struct{}

func (t *Telemetry) Shutdown(_ context.Context) {}

func initializeTelemetry(_ *zap.Logger) (*Telemetry, error) {
	return &Telemetry{}, nil
}
