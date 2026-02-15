//go:generate oapi-codegen -generate types,server -package api -o api.gen.go ../../api/openapi.yaml
//go:generate mockgen -source=api.gen.go -destination=server.gen.go -package=api

package api

import (
	"fcstask/internal/server"

	handler "fcstask/internal/server/handler"

	"github.com/labstack/echo/v4"
)

func RegisterHandlers(e *echo.Echo, apiServer *server.Server) {
	e.GET("/api/courses", handler.GetCoursesHandler)
	e.GET("/api/coursses/:courseId", handler.GetCourseHandler)
	e.POST("/api/courses", handler.CreateCourseHandler)
	e.PUT("/api/courses/:courseId", handler.UpdateCourseHandler)

	e.GET("/api/courses/:courseId/board", handler.GetCourseBoardHandler)
}
