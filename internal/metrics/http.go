package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type HTTPMetrics struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	ResponseSize    *prometheus.HistogramVec
	InFlight        prometheus.Gauge
}

var httpLatencyBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

var httpResponseSizeBuckets = []float64{
	100, 500, 1_000, 5_000, 10_000, 50_000, 100_000, 500_000, 1_000_000, 5_000_000, 10_000_000,
}

func newHTTPMetrics(reg prometheus.Registerer) *HTTPMetrics {
	factory := promauto.With(reg)

	return &HTTPMetrics{
		RequestsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests processed.",
			},
			[]string{"method", "route", "status"},
		),
		RequestDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: Namespace,
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "Latency of HTTP requests in seconds.",
				Buckets:   httpLatencyBuckets,
			},
			[]string{"method", "route"},
		),
		ResponseSize: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: Namespace,
				Subsystem: "http",
				Name:      "response_size_bytes",
				Help:      "Size of HTTP response bodies in bytes.",
				Buckets:   httpResponseSizeBuckets,
			},
			[]string{"method", "route"},
		),
		InFlight: factory.NewGauge(
			prometheus.GaugeOpts{
				Namespace: Namespace,
				Subsystem: "http",
				Name:      "requests_in_flight",
				Help:      "Number of HTTP requests currently being processed.",
			},
		),
	}
}
