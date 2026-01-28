package server

import (
	"fcstask/internal/server/handler"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (s *APIServer) PostV1Echo(ctx echo.Context) error {
	return handler.Echo(ctx)
}

func (s *APIServer) CreateUser(ctx echo.Context) error {
	return handler.CreateUserHandler(s.userRepo, ctx)
}

func (s *APIServer) GetUserByID(ctx echo.Context, id int64) error {
	return s.GetUserByID(ctx, id)
}

func (s *APIServer) GetUserByUsername(ctx echo.Context, username string) error {
	return s.GetUserByUsername(ctx, username)
}

func (s *APIServer) GetUserByEmail(ctx echo.Context, email openapi_types.Email) error {
	return s.GetUserByEmail(ctx, email)
}
