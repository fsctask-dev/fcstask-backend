package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type AuthService struct {
	userRepo    repo.IUserRepo
	sessionRepo repo.SessionRepositoryInterface
}

func NewAuthService(userRepo repo.IUserRepo, sessionRepo repo.SessionRepositoryInterface) *AuthService {
	return &AuthService{userRepo: userRepo, sessionRepo: sessionRepo}
}

type SignUpInput struct {
	Email     string
	Username  string
	Password  string
	TgUID     *int64
	IP        string
	UserAgent string
}

type SignInInput struct {
	Email     *string
	Username  *string
	Password  string
	IP        string
	UserAgent string
}

type AuthResult struct {
	User    *models.User
	Session *models.Session
}

func (s *AuthService) SignUp(ctx context.Context, input SignUpInput) (*AuthResult, error) {
	if input.Username == "" || input.Password == "" || input.Email == "" {
		return nil, BadRequest("Username, password and email are required")
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

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, Internal("Failed to hash password", err)
	}

	user := &models.User{
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: string(hash),
		TgUID:        input.TgUID,
		UserID:       uuid.New(),
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		if col := UniqueConstraintColumn(err); col != "" {
			return nil, Conflict("User with this " + col + " already exists")
		}
		return nil, Internal("Failed to create user", err)
	}

	session, err := s.createSession(ctx, user.ID, input.IP, input.UserAgent)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, Session: session}, nil
}

func (s *AuthService) SignIn(ctx context.Context, input SignInInput) (*AuthResult, error) {
	if input.Password == "" {
		return nil, BadRequest("Password is required")
	}

	if input.Email == nil && input.Username == nil {
		return nil, BadRequest("Email or username is required")
	}

	var user *models.User
	var err error
	if input.Email != nil {
		user, err = s.userRepo.GetUserByEmail(ctx, *input.Email)
	} else {
		user, err = s.userRepo.GetUserByUsername(ctx, *input.Username)
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, Unauthorized("Invalid credentials")
		}
		return nil, Internal("Failed to find user", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, Unauthorized("Invalid credentials")
	}

	session, err := s.createSession(ctx, user.ID, input.IP, input.UserAgent)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, Session: session}, nil
}

func (s *AuthService) GetMe(ctx context.Context, user *models.User) (string, string, error) {
	if user == nil {
		return "", "", Unauthorized("Not authenticated")
	}
	return buildInitials(user), "user", nil
}

func (s *AuthService) SignOut(ctx context.Context, session *models.Session) error {
	if session == nil {
		return Unauthorized("Not authenticated")
	}
	if err := s.sessionRepo.DeleteSession(ctx, session.ID); err != nil {
		return Internal("Failed to delete session", err)
	}
	return nil
}

func (s *AuthService) createSession(ctx context.Context, userID uuid.UUID, ip, userAgent string) (*models.Session, error) {
	session := &models.Session{
		UserID:    userID,
		IP:        ip,
		UserAgent: userAgent,
	}
	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		return nil, Internal("Failed to create session", err)
	}
	return session, nil
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
