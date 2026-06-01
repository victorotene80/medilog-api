package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "medilog",
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "HTTP request duration in seconds bucketed by latency",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "route", "status"})
)

// PrometheusMetrics records request duration / status / method / route.
// Remove this file and re-enable middleware/telemetry.go to switch back to OTel.
type PrometheusMetrics struct{}

func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{}
}

func (m *PrometheusMetrics) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		route := r.URL.Path
		if rctx := chi.RouteContext(r.Context()); rctx != nil && rctx.RoutePattern() != "" {
			route = rctx.RoutePattern()
		}

		httpDuration.WithLabelValues(
			r.Method,
			route,
			strconv.Itoa(rw.status),
		).Observe(time.Since(start).Seconds())
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
