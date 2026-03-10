package middleware

import (
	"net/http"
	"strconv"
	"time"

	"fcstask/internal/metrics"
)

type MetricsMiddleware struct {
	client metrics.MetricsClient
}

func NewMetricsMiddleware(client metrics.MetricsClient) *MetricsMiddleware {
	return &MetricsMiddleware{client: client}
}

type ResponseWriter struct {
	http.ResponseWriter
	status int
}

func (m *MetricsMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now()

		rw := &ResponseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(timeStart).Seconds()
		path := r.URL.Path
		method := r.Method
		status := strconv.Itoa(rw.status)

		m.client.Inc("http_requests_total", map[string]string{
			"path":   path,
			"method": method,
			"status": status,
		})

		m.client.Histogram("http_request_duration_seconds", duration, map[string]string{
			"course": path,
			"method": method,
		})
	})
}

func (rw *ResponseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
