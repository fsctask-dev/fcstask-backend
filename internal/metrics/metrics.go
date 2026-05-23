package metrics

import (
	"net/http"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
)

func EchoPrometheus(e *echo.Echo) {
	e.Use(echoprometheus.NewMiddleware("echo"))

	e.GET("/metrics", echoprometheus.NewHandler())
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
}
