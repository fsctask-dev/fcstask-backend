package controller

import (
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"fcstask-backend/internal/api"
	"fcstask-backend/internal/handler"
)

type APIController struct {
	authHandler    *handler.AuthHandler
	userHandler    *handler.UserHandler
	sessionHandler *handler.SessionHandler
	courseHandler  *handler.CourseHandler
}

func NewAPIController(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	sessionHandler *handler.SessionHandler,
	courseHandler *handler.CourseHandler,
) *APIController {
	return &APIController{
		authHandler:    authHandler,
		userHandler:    userHandler,
		sessionHandler: sessionHandler,
		courseHandler:  courseHandler,
	}
}

func (c *APIController) PostV1Echo(ctx echo.Context) error {
	return handler.Echo(ctx)
}

func (c *APIController) CreateUser(ctx echo.Context) error {
	return c.userHandler.CreateUser(ctx)
}

func (c *APIController) GetUserByID(ctx echo.Context, id openapi_types.UUID) error {
	return c.userHandler.GetUserByID(ctx, id)
}

func (c *APIController) GetUserByUsername(ctx echo.Context, username string) error {
	return c.userHandler.GetUserByUsername(ctx, username)
}

func (c *APIController) GetUserByEmail(ctx echo.Context, email openapi_types.Email) error {
	return c.userHandler.GetUserByEmail(ctx, email)
}

func (c *APIController) SignUp(ctx echo.Context) error {
	return c.authHandler.SignUp(ctx)
}

func (c *APIController) SignIn(ctx echo.Context) error {
	return c.authHandler.SignIn(ctx)
}

func (c *APIController) SignOut(ctx echo.Context) error {
	return c.authHandler.SignOut(ctx)
}

func (c *APIController) GetMe(ctx echo.Context) error {
	return c.authHandler.GetMe(ctx)
}

func (c *APIController) GetSessions(ctx echo.Context, params api.GetSessionsParams) error {
	return c.sessionHandler.GetSessions(ctx, params)
}

func (c *APIController) GetUsersWithSessions(ctx echo.Context, params api.GetUsersWithSessionsParams) error {
	return c.sessionHandler.GetUsersWithSessions(ctx, params)
}

func (c *APIController) RegisterCourseRoutes(e *echo.Echo) {
	e.GET("/api/courses", c.courseHandler.GetCourses)
	e.POST("/api/courses", c.courseHandler.CreateCourse)
	e.GET("/api/courses/:courseId", c.courseHandler.GetCourse)
	e.PUT("/api/courses/:courseId", c.courseHandler.UpdateCourse)
	e.GET("/api/courses/:courseId/board", c.courseHandler.GetCourseBoard)
}
