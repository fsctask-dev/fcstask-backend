package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type UserService struct {
	userRepo repo.IUserRepo
}

func NewUserService(userRepo repo.IUserRepo) *UserService {
	return &UserService{userRepo: userRepo}
}

type CreateUserInput struct {
	Email     string
	Username  string
	FirstName *string
	LastName  *string
	TgUID     *int64
	UserID    uuid.UUID
}

func (s *UserService) CreateUser(ctx context.Context, input CreateUserInput) (*models.User, error) {
	if input.Email == "" || input.Username == "" || input.UserID == uuid.Nil {
		return nil, BadRequest("Email, username and user_id are required")
	}

	if exists, err := s.userRepo.ExistsUserByEmail(ctx, input.Email); err != nil {
		return nil, Internal("Failed to check email uniqueness", err)
	} else if exists {
		return nil, Conflict("User with this email already exists")
	}

	if exists, err := s.userRepo.ExistsUserByUsername(ctx, input.Username); err != nil {
		return nil, Internal("Failed to check username uniqueness", err)
	} else if exists {
		return nil, Conflict("User with this username already exists")
	}

	if input.TgUID != nil && *input.TgUID != 0 {
		existingUser, err := s.userRepo.GetUserByTgUID(ctx, *input.TgUID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, Internal("Failed to check tg_uid uniqueness", err)
		}
		if existingUser != nil {
			return nil, Conflict("User with this tg_uid already exists")
		}
	}

	user := &models.User{
		Email:     input.Email,
		Username:  input.Username,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		TgUID:     input.TgUID,
		UserID:    input.UserID,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		if col := UniqueConstraintColumn(err); col != "" {
			return nil, Conflict("User with this " + col + " already exists")
		}
		return nil, Internal("Failed to create user", err)
	}

	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NotFound("User not found")
		}
		return nil, Internal("Failed to get user", err)
	}
	return user, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NotFound("User not found")
		}
		return nil, Internal("Failed to get user", err)
	}
	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NotFound("User not found")
		}
		return nil, Internal("Failed to get user", err)
	}
	return user, nil
}

func (s *UserService) GetUsersWithSessions(ctx context.Context, limit, offset int) ([]models.User, int64, error) {
	limit, offset, err := ParsePagination(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.userRepo.CountUsersWithSessions(ctx)
	if err != nil {
		return nil, 0, Internal("Failed to count users", err)
	}

	users, err := s.userRepo.GetUsersWithSessions(ctx, limit, offset)
	if err != nil {
		return nil, 0, Internal("Failed to get users with sessions", err)
	}

	return users, total, nil
}

func UniqueConstraintColumn(err error) string {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return ""
	}

	parts := strings.SplitN(pgErr.ConstraintName, "_", 3)
	if len(parts) == 3 {
		return parts[2]
	}
	return pgErr.ConstraintName
}
