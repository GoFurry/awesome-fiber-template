package metrics

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type HTTPMetrics struct {
	HttpRequestsTotal   *prometheus.CounterVec
	HttpRequestDuration *prometheus.HistogramVec
	HttpActiveRequests  prometheus.Gauge
}

func New(namespace string) *HTTPMetrics {
	namespace = sanitizeNamespace(namespace)

	return &HTTPMetrics{
		HttpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),

		HttpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request latency",
				Buckets:   []float64{0.05, 0.1, 0.2, 0.3, 0.5, 1, 1.5, 2, 3, 5},
			},
			[]string{"method", "path"},
		),

		HttpActiveRequests: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "active_requests",
				Help:      "Number of active HTTP requests",
			},
		),
	}
}

func sanitizeNamespace(namespace string) string {
	namespace = strings.TrimSpace(namespace)
	namespace = strings.ReplaceAll(namespace, "-", "_")
	if namespace == "" {
		return "fiberx"
	}
	return namespace
}
