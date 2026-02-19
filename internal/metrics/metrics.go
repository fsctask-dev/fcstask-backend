package metrics

import (
	"net/http"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
)

func EchoPrometheus(e *echo.Echo) {
	e.Use(echoprometheus.NewMiddleware("echo"))

	e.GET("/metrics", echoprometheus.NewHandler())
	e.GET("/test/200", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK") // 200
	})
	e.GET("/test/500", func(c echo.Context) error {
		return c.String(http.StatusInternalServerError, "Internal Server Error") // 500
	})
}
