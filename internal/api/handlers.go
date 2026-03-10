package api

import "github.com/labstack/echo/v4"

// Handler interface for the API
type ServerInterface interface {
	PostV1Echo(ctx echo.Context) error
}

func RegisterHandlers(e *echo.Echo, server ServerInterface) {
	e.POST("/v1/echo", server.PostV1Echo)
}
