package metrics

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

func EchoMiddleware(m *HTTPMetrics) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			m.InFlight.Inc()
			start := time.Now()

			err := next(c)

			m.InFlight.Dec()

			route := c.Path()
			if route == "" {
				route = "unmatched"
			}
			method := c.Request().Method
			status := strconv.Itoa(c.Response().Status)

			m.RequestsTotal.WithLabelValues(method, route, status).Inc()
			m.RequestDuration.WithLabelValues(method, route).Observe(time.Since(start).Seconds())
			m.ResponseSize.WithLabelValues(method, route).Observe(float64(c.Response().Size))

			return err
		}
	}
}
