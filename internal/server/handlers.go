package server

import (
	"fcstask/internal/api"
	"fcstask/internal/server/handler"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (s *APIServer) PostV1Echo(ctx echo.Context) error {
	return handler.Echo(ctx)
}

func (s *APIServer) CreateUser(ctx echo.Context) error {
	return handler.CreateUserHandler(s.userRepo, ctx)
}

func (s *APIServer) GetUserByID(ctx echo.Context, id openapi_types.UUID) error {
	return handler.GetUserByIDHandler(s.userRepo, ctx, uuid.UUID(id))
}

func (s *APIServer) GetUserByUsername(ctx echo.Context, username string) error {
	return handler.GetUserByUsernameHandler(s.userRepo, ctx, username)
}

func (s *APIServer) GetUserByEmail(ctx echo.Context, email openapi_types.Email) error {
	return handler.GetUserByEmailHandler(s.userRepo, ctx, string(email))
}

func (s *APIServer) SignUp(ctx echo.Context) error {
	return handler.SignUpHandler(s.userRepo, s.sessionRepo, ctx)
}

func (s *APIServer) SignIn(ctx echo.Context) error {
	return handler.SignInHandler(s.userRepo, s.sessionRepo, ctx)
}

func (s *APIServer) SignOut(ctx echo.Context) error {
	return handler.SignOutHandler(s.sessionRepo, ctx)
}

func (s *APIServer) GetMe(ctx echo.Context) error {
	return handler.GetMeHandler(s.userRepo, s.sessionRepo, ctx)
}

func (s *APIServer) GetSessions(ctx echo.Context, params api.GetSessionsParams) error {
	return handler.GetSessionsHandler(s.sessionRepo, ctx, params)
}

func (s *APIServer) GetUsersWithSessions(ctx echo.Context, params api.GetUsersWithSessionsParams) error {
	return handler.GetUsersWithSessionsHandler(s.userRepo, ctx, params)
}
