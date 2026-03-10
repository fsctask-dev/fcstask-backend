package handler

import (
	"errors"
	models "fcstask/internal/db/model"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"gorm.io/gorm"

	"fcstask/internal/api"
	"fcstask/internal/db/repo"
)

func CreateUserHandler(userRepo repo.UserRepositoryInterface, ctx echo.Context) error {
	var req api.CreateUserRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}

	if req.Email == "" || req.Username == "" || req.UserId == uuid.Nil {
		return badRequest(ctx, "Email, username and user_id are required")
	}

	if exists, err := userRepo.ExistsByEmail(ctx.Request().Context(), string(req.Email)); err != nil {
		return internalError(ctx, "Failed to check email uniqueness")
	} else if exists {
		return conflict(ctx, "User with this email already exists")
	}

	if exists, err := userRepo.ExistsByUsername(ctx.Request().Context(), req.Username); err != nil {
		return internalError(ctx, "Failed to check username uniqueness")
	} else if exists {
		return conflict(ctx, "User with this username already exists")
	}

	// Проверяем уникальность tg_uid если передан
	if req.TgUid != nil && *req.TgUid != 0 {
		existingUser, err := userRepo.GetByTgUID(ctx.Request().Context(), *req.TgUid)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return internalError(ctx, "Failed to check tg_uid uniqueness")
		}
		if existingUser != nil {
			return conflict(ctx, "User with this tg_uid already exists")
		}
	}

	user := models.User{
		Email:     string(req.Email),
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		TgUID:     req.TgUid,
		UserID:    uuid.UUID(req.UserId),
	}

	if err := userRepo.Create(ctx.Request().Context(), &user); err != nil {
		if col := uniqueConstraintColumn(err); col != "" {
			return conflict(ctx, "User with this "+col+" already exists")
		}
		return internalError(ctx, "Failed to create user")
	}

	userResp := api.User{
		Id:        user.ID,
		Email:     req.Email,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		TgUid:     user.TgUID,
		UserId:    openapi_types.UUID(user.UserID),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return ctx.JSON(http.StatusCreated, userResp)
}

func GetUserByIDHandler(userRepo repo.UserRepositoryInterface, ctx echo.Context, id uuid.UUID) error {
	user, err := userRepo.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apiError(ctx, http.StatusNotFound, "not_found", "User not found")
		}
		return internalError(ctx, "Failed to get user")
	}

	return ctx.JSON(http.StatusOK, userToAPI(user))
}

func GetUserByUsernameHandler(userRepo repo.UserRepositoryInterface, ctx echo.Context, username string) error {
	user, err := userRepo.GetByUsername(ctx.Request().Context(), username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apiError(ctx, http.StatusNotFound, "not_found", "User not found")
		}
		return internalError(ctx, "Failed to get user")
	}

	return ctx.JSON(http.StatusOK, userToAPI(user))
}

func GetUserByEmailHandler(userRepo repo.UserRepositoryInterface, ctx echo.Context, email string) error {
	user, err := userRepo.GetByEmail(ctx.Request().Context(), email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apiError(ctx, http.StatusNotFound, "not_found", "User not found")
		}
		return internalError(ctx, "Failed to get user")
	}

	return ctx.JSON(http.StatusOK, userToAPI(user))
}

func userToAPI(user *models.User) api.User {
	return api.User{
		Id:        user.ID,
		Email:     openapi_types.Email(user.Email),
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		TgUid:     user.TgUID,
		UserId:    openapi_types.UUID(user.UserID),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
