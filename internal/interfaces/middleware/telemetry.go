// OTel telemetry middleware (traces + metrics).
// Remove the //go:build ignore line above to re-enable.
// ================================================================================

package middleware

import (
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/victorotene80/medilog-api"

type TelemetryMiddleware struct {
	tracer         trace.Tracer
	duration       metric.Float64Histogram
	activeRequests metric.Int64UpDownCounter
}

func NewTelemetryMiddleware() (*TelemetryMiddleware, error) {
	tracer := otel.Tracer(instrumentationName)
	meter := otel.Meter(instrumentationName)

	duration, err := meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("Duration of HTTP server requests in milliseconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry middleware: create duration histogram: %w", err)
	}

	activeRequests, err := meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithDescription("Number of in-flight HTTP requests"),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry middleware: create active requests counter: %w", err)
	}

	return &TelemetryMiddleware{
		tracer:         tracer,
		duration:       duration,
		activeRequests: activeRequests,
	}, nil
}

func (m *TelemetryMiddleware) Handle(next http.Handler) http.Handler {
	propagator := otel.GetTextMapPropagator()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		route := r.URL.Path
		method := r.Method

		commonAttrs := []attribute.KeyValue{
			semconv.HTTPRequestMethodKey.String(method),
			semconv.URLPath(route),
		}

		spanName := fmt.Sprintf("%s %s", method, route)
		ctx, span := m.tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.ServerAddress(r.Host),
				semconv.NetworkProtocolVersion(r.Proto),
			),
		)
		defer span.End()

		m.activeRequests.Add(ctx, 1, metric.WithAttributes(commonAttrs...))
		defer m.activeRequests.Add(ctx, -1, metric.WithAttributes(commonAttrs...))

		rw := &telemetryStatusRecorder{ResponseWriter: w, status: http.StatusOK}

		start := time.Now()
		next.ServeHTTP(rw, r.WithContext(ctx))
		elapsedMs := float64(time.Since(start).Milliseconds())

		statusAttrs := append(commonAttrs,
			semconv.HTTPResponseStatusCode(rw.status),
		)

		span.SetAttributes(semconv.HTTPResponseStatusCode(rw.status))

		m.duration.Record(ctx, elapsedMs, metric.WithAttributes(statusAttrs...))
	})
}

type telemetryStatusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *telemetryStatusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
