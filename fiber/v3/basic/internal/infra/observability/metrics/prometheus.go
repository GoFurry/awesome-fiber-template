package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "awesome_template",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "awesome_template",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request latency",
			Buckets:   []float64{0.05, 0.1, 0.2, 0.3, 0.5, 1, 1.5, 2, 3, 5},
		},
		[]string{"method", "path"},
	)

	HttpActiveRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "awesome_template",
			Subsystem: "http",
			Name:      "active_requests",
			Help:      "Number of active HTTP requests",
		},
	)
)
