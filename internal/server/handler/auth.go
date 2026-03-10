package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"fcstask/internal/api"
	models "fcstask/internal/db/model"
	"fcstask/internal/db/repo"
)

func SignUpHandler(userRepo repo.UserRepositoryInterface, sessionRepo repo.SessionRepositoryInterface, ctx echo.Context) error {
	var req api.SignUpRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}

	if req.Username == "" || req.Password == "" || req.Email == "" {
		return badRequest(ctx, "Username, password and email are required")
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

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return internalError(ctx, "Failed to hash password")
	}

	user := models.User{
		Email:        string(req.Email),
		Username:     req.Username,
		PasswordHash: string(hash),
		TgUID:        req.TgUid,
		UserID:       uuid.New(),
	}

	if err := userRepo.Create(ctx.Request().Context(), &user); err != nil {
		if col := uniqueConstraintColumn(err); col != "" {
			return conflict(ctx, "User with this "+col+" already exists")
		}
		return internalError(ctx, "Failed to create user")
	}

	session := models.Session{
		UserID:    user.ID,
		IP:        ctx.RealIP(),
		UserAgent: ctx.Request().UserAgent(),
	}
	if err := sessionRepo.Create(ctx.Request().Context(), &session); err != nil {
		return internalError(ctx, "Failed to create session")
	}

	return ctx.JSON(http.StatusCreated, api.AuthResponse{
		SessionToken: openapi_types.UUID(session.ID),
		User: api.User{
			Id:        user.ID,
			Email:     openapi_types.Email(user.Email),
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			TgUid:     user.TgUID,
			UserId:    openapi_types.UUID(user.UserID),
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	})
}

func SignInHandler(userRepo repo.UserRepositoryInterface, sessionRepo repo.SessionRepositoryInterface, ctx echo.Context) error {
	var req api.SignInRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}

	if req.Password == "" {
		return badRequest(ctx, "Password is required")
	}

	if req.Email == nil && req.Username == nil {
		return badRequest(ctx, "Email or username is required")
	}

	var user *models.User
	var err error

	if req.Email != nil {
		user, err = userRepo.GetByEmail(ctx.Request().Context(), string(*req.Email))
	} else {
		user, err = userRepo.GetByUsername(ctx.Request().Context(), *req.Username)
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return unauthorized(ctx, "Invalid credentials")
		}
		return internalError(ctx, "Failed to find user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return unauthorized(ctx, "Invalid credentials")
	}

	session := models.Session{
		UserID:    user.ID,
		IP:        ctx.RealIP(),
		UserAgent: ctx.Request().UserAgent(),
	}
	if err := sessionRepo.Create(ctx.Request().Context(), &session); err != nil {
		return internalError(ctx, "Failed to create session")
	}

	return ctx.JSON(http.StatusOK, api.AuthResponse{
		SessionToken: openapi_types.UUID(session.ID),
		User: api.User{
			Id:        user.ID,
			Email:     openapi_types.Email(user.Email),
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			TgUid:     user.TgUID,
			UserId:    openapi_types.UUID(user.UserID),
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	})
}

const UserContextKey = "user"
const SessionContextKey = "session"

func GetMeHandler(userRepo repo.UserRepositoryInterface, sessionRepo repo.SessionRepositoryInterface, ctx echo.Context) error {
	user, ok := ctx.Get(UserContextKey).(*models.User)
	if !ok || user == nil {
		return unauthorized(ctx, "Not authenticated")
	}

	initials := buildInitials(user)

	return ctx.JSON(http.StatusOK, api.MeResponse{
		Username: user.Username,
		Initials: initials,
		Role:     "user",
	})
}

func SignOutHandler(sessionRepo repo.SessionRepositoryInterface, ctx echo.Context) error {
	session, ok := ctx.Get(SessionContextKey).(*models.Session)
	if !ok || session == nil {
		return unauthorized(ctx, "Not authenticated")
	}

	if err := sessionRepo.Delete(ctx.Request().Context(), session.ID); err != nil {
		return internalError(ctx, "Failed to delete session")
	}

	return ctx.NoContent(http.StatusNoContent)
}

func buildInitials(user *models.User) string {
	var parts []string
	if user.FirstName != nil && *user.FirstName != "" {
		parts = append(parts, string([]rune(*user.FirstName)[0:1]))
	}
	if user.LastName != nil && *user.LastName != "" {
		parts = append(parts, string([]rune(*user.LastName)[0:1]))
	}
	if len(parts) == 0 {
		r := []rune(user.Username)
		if len(r) >= 2 {
			return strings.ToUpper(fmt.Sprintf("%c%c", r[0], r[1]))
		}
		if len(r) == 1 {
			return strings.ToUpper(string(r))
		}
		return "?"
	}
	return strings.ToUpper(strings.Join(parts, ""))
}
