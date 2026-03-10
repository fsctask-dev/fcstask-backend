package middleware

import (
	"net/http"

	"fcstask/internal/metrics"
)

type InflightMetricsMiddleware struct {
	client metrics.MetricsClient
}

func NewInflightMiddleware(client metrics.MetricsClient) *InflightMetricsMiddleware {
	return &InflightMetricsMiddleware{client: client}
}

func (im *InflightMetricsMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		im.client.Gauge("http_requests_in_flight", 1, nil)

		defer im.client.Gauge("http_requests_in_flight", -1, nil)

		next.ServeHTTP(w, r)
	})
}
