package handler

import (
	"errors"
	models "fcstask/internal/db/model"
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"gorm.io/gorm"

	"fcstask/internal/api"
	"fcstask/internal/db/repo"
)

func CreateUserHandler(userRepo repo.UserRepositoryInterface, ctx echo.Context) error {
	var req api.CreateUserRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" || req.Username == "" || req.UserId == 0 {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Email, username and user_id are required",
		})
	}

	if exists, err := userRepo.ExistsByEmail(ctx.Request().Context(), string(req.Email)); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to check email uniqueness",
		})
	} else if exists {
		return ctx.JSON(http.StatusConflict, map[string]string{
			"error": "User with this email already exists",
		})
	}

	if exists, err := userRepo.ExistsByUsername(ctx.Request().Context(), req.Username); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to check username uniqueness",
		})
	} else if exists {
		return ctx.JSON(http.StatusConflict, map[string]string{
			"error": "User with this username already exists",
		})
	}

	// Проверяем уникальность tg_uid если передан
	if req.TgUid != nil && *req.TgUid != 0 {
		existingUser, err := userRepo.GetByTgUID(ctx.Request().Context(), *req.TgUid)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to check tg_uid uniqueness",
			})
		}
		if existingUser != nil {
			return ctx.JSON(http.StatusConflict, map[string]string{
				"error": "User with this tg_uid already exists",
			})
		}
	}

	user := models.User{
		Email:     string(req.Email),
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		TgUID:     req.TgUid,
		UserID:    req.UserId,
	}

	if err := userRepo.Create(ctx.Request().Context(), &user); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user",
		})
	}

	userResp := api.User{
		Id:        int64(user.ID),
		Email:     req.Email,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		TgUid:     user.TgUID,
		UserId:    user.UserID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return ctx.JSON(http.StatusCreated, userResp)
}

func GetUserByIDHandler(userRepo repo.UserRepositoryInterface, ctx echo.Context, id int64) error {
	user, err := userRepo.GetByID(ctx.Request().Context(), uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "User not found",
			})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get user",
		})
	}

	userResp := api.User{
		Id:        int64(user.ID),
		Email:     openapi_types.Email(user.Email),
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		TgUid:     user.TgUID,
		UserId:    user.UserID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return ctx.JSON(http.StatusOK, userResp)
}

func GetUserByUsernameHandler(userRepo repo.UserRepositoryInterface, ctx echo.Context, username string) error {
	user, err := userRepo.GetByUsername(ctx.Request().Context(), username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "User not found",
			})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get user",
		})
	}

	userResp := api.User{
		Id:        int64(user.ID),
		Email:     openapi_types.Email(user.Email),
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		TgUid:     user.TgUID,
		UserId:    user.UserID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return ctx.JSON(http.StatusOK, userResp)
}

func GetUserByEmailHandler(userRepo repo.UserRepositoryInterface, ctx echo.Context, email string) error {
	user, err := userRepo.GetByEmail(ctx.Request().Context(), email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "User not found",
			})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get user",
		})
	}

	userResp := api.User{
		Id:        int64(user.ID),
		Email:     openapi_types.Email(user.Email),
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		TgUid:     user.TgUID,
		UserId:    user.UserID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return ctx.JSON(http.StatusOK, userResp)
}
