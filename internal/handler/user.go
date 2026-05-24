package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"fcstask-backend/internal/api"
	"fcstask-backend/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) CreateUser(ctx echo.Context) error {
	var req api.CreateUserRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}

	user, err := h.userService.CreateUser(ctx.Request().Context(), service.CreateUserInput{
		Email:     string(req.Email),
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		TgUID:     req.TgUid,
		UserID:    uuid.UUID(req.UserId),
	})
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, userToAPI(user))
}

func (h *UserHandler) GetUserByID(ctx echo.Context, id openapi_types.UUID) error {
	user, err := h.userService.GetUserByID(ctx.Request().Context(), uuid.UUID(id))
	if err != nil {
		return serviceError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, userToAPI(user))
}

func (h *UserHandler) GetUserByUsername(ctx echo.Context, username string) error {
	user, err := h.userService.GetUserByUsername(ctx.Request().Context(), username)
	if err != nil {
		return serviceError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, userToAPI(user))
}

func (h *UserHandler) GetUserByEmail(ctx echo.Context, email openapi_types.Email) error {
	user, err := h.userService.GetUserByEmail(ctx.Request().Context(), string(email))
	if err != nil {
		return serviceError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, userToAPI(user))
}
