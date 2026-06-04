package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"fcstask-backend/internal/config"
	"fcstask-backend/internal/db/model"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
	"fcstask-backend/internal/mailer"
	"fcstask-backend/internal/metrics"
)

// AuthService owns email/password registration (with email verification) and
// the session lifecycle (sign in / out, "me").
type AuthService struct {
	userRepo              repo.IUserRepo
	sessionRepo           repo.ISessionRepository
	emailRegistrationRepo repo.IEmailRegistrationRepo
	oauthRepo             repo.IOAuthIdentityRepo

	mailer mailer.Mailer

	emailRegistrationConfig config.EmailRegistrationConfig

	authMetrics    *metrics.AuthMetrics
	sessionMetrics *metrics.SessionMetrics
}

func NewAuthService(
	userRepo repo.IUserRepo,
	sessionRepo repo.ISessionRepository,
	emailRegistrationRepo repo.IEmailRegistrationRepo,
	oauthRepo repo.IOAuthIdentityRepo,
	m mailer.Mailer,
	emailRegistrationConfig config.EmailRegistrationConfig,
) *AuthService {
	return &AuthService{
		userRepo:                userRepo,
		sessionRepo:             sessionRepo,
		emailRegistrationRepo:   emailRegistrationRepo,
		oauthRepo:               oauthRepo,
		mailer:                  m,
		emailRegistrationConfig: emailRegistrationConfig,
	}
}

func (s *AuthService) WithMetrics(auth *metrics.AuthMetrics, session *metrics.SessionMetrics) *AuthService {
	s.authMetrics = auth
	s.sessionMetrics = session
	return s
}

type SignUpInput struct {
	Email     string
	Username  string
	Password  string
	FirstName *string
	LastName  *string
}

type SignUpVerifyInput struct {
	Token       uuid.UUID
	Code        string
	MaxAttempts int
	IP          string
	UserAgent   string
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

// SignUp begins an email/password registration. The user row is not created
// yet — instead a pending EmailRegistration is stored and a verification code
// is emailed. The caller finishes the flow via SignUpVerify. The returned
// EmailRegistration's ID is the verification_token the client passes back.
func (s *AuthService) SignUp(ctx context.Context, input SignUpInput) (result *model.EmailRegistration, err error) {
	defer func() { s.authMetrics.IncSignup(authOutcomeFromError(err)) }()

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

	// Drop any previous outstanding registration for this email so the latest
	// attempt wins (mirrors PasswordResetRequest's per-user cleanup).
	if err := s.emailRegistrationRepo.DeleteByEmail(ctx, input.Email); err != nil {
		return nil, Internal("Failed to delete previous registration", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, Internal("Failed to hash password", err)
	}

	code, err := generateCode()
	if err != nil {
		return nil, Internal("Failed to generate verification code", err)
	}
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return nil, Internal("Failed to hash verification code", err)
	}

	now := time.Now()
	result = &models.EmailRegistration{
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: string(passwordHash),
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		CodeHash:     string(codeHash),
		LastSentAt:   now,
		ExpiresAt:    now.Add(s.emailRegistrationConfig.TTL),
	}
	if err := mailer.SendEmailConfirmation(s.mailer, ctx, result, code); err != nil {
		return nil, Internal("failed to send verification email", err)
	}
	if err := s.emailRegistrationRepo.Create(ctx, result); err != nil {
		if col := UniqueConstraintColumn(err); col != "" {
			return nil, Conflict("User with this " + col + " already exists")
		}
		return nil, Internal("failed to create registration", err)
	}
	return result, nil
}

// SignUpVerify finishes a registration started by SignUp: it validates the
// emailed code, creates the user, opens a session and discards the pending
// registration.
func (s *AuthService) SignUpVerify(ctx context.Context, input SignUpVerifyInput) (result *AuthResult, err error) {
	defer func() { s.authMetrics.IncSignup(authOutcomeFromError(err)) }()

	reg, err := s.emailRegistrationRepo.GetByID(ctx, input.Token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NotFound("verification token not found")
		}
		return nil, Internal("Failed to get registration", err)
	}

	now := time.Now()
	if reg.ExpiresAt.Before(now) {
		return nil, Unauthorized("verification token has expired")
	}

	if reg.Attempts > input.MaxAttempts {
		return nil, Unauthorized("number of attempts exceeded")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(reg.CodeHash), []byte(input.Code)); err != nil {
		reg.Attempts++
		if err := s.emailRegistrationRepo.Update(ctx, reg); err != nil {
			return nil, Internal("failed to update registration", err)
		}
		return nil, Unauthorized("verification codes do not match")
	}

	// Re-check uniqueness: another account may have claimed the email or
	// username between SignUp and SignUpVerify.
	user := &models.User{
		Email:        reg.Email,
		Username:     reg.Username,
		PasswordHash: reg.PasswordHash,
		FirstName:    reg.FirstName,
		LastName:     reg.LastName,
		UserID:       uuid.New(),
	}
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, Internal("Failed to create user", err)
	}

	identity, err := s.oauthRepo.GetByEmailRegistrationID(ctx, reg.ID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, Internal("Failed to fetch identity", err)
		} else {
			identity.EmailRegistrationID = uuid.Nil
			identity.UserID = user.ID
			if err := s.oauthRepo.Update(ctx, identity); err != nil {
				return nil, Internal("Failed to update identity", err)
			}
		}
	}

	if err := s.emailRegistrationRepo.Delete(ctx, reg.ID); err != nil {
		return nil, Internal("failed to delete registration", err)
	}

	session, err := openSession(ctx, s.sessionRepo, s.sessionMetrics, user.ID, input.IP, input.UserAgent)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, Session: session}, nil
}

// SignUpResend re-issues the verification code for a pending registration,
// identified by the verification_token returned from SignUp.
func (s *AuthService) SignUpResend(ctx context.Context, token uuid.UUID) (result *model.EmailRegistration, err error) {
	result, err = s.emailRegistrationRepo.GetByID(ctx, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NotFound("verification token not found")
		}
		return nil, Internal("Failed to get registration", err)
	}

	now := time.Now()
	if result.ExpiresAt.Before(now) {
		return nil, Unauthorized("verification token has expired")
	}

	code, err := generateCode()
	if err != nil {
		return nil, Internal("Failed to generate verification code", err)
	}
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return nil, Internal("Failed to hash verification code", err)
	}

	result.CodeHash = string(codeHash)
	result.LastSentAt = now
	result.ExpiresAt = now.Add(s.emailRegistrationConfig.TTL)
	result.Attempts = 0

	if err := mailer.SendEmailConfirmation(s.mailer, ctx, result, code); err != nil {
		return nil, Internal("failed to send verification email", err)
	}
	if err := s.emailRegistrationRepo.Update(ctx, result); err != nil {
		return nil, Internal("failed to update registration", err)
	}
	return result, nil
}

func (s *AuthService) SignIn(ctx context.Context, input SignInInput) (result *AuthResult, err error) {
	defer func() { s.authMetrics.IncSignIn(authOutcomeFromError(err)) }()

	if input.Password == "" {
		return nil, BadRequest("Password is required")
	}

	if input.Email == nil && input.Username == nil {
		return nil, BadRequest("Email or username is required")
	}

	var user *models.User
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

	session, err := openSession(ctx, s.sessionRepo, s.sessionMetrics, user.ID, input.IP, input.UserAgent)
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
	s.authMetrics.IncSignOut()
	s.sessionMetrics.IncRevoked(metrics.SessionRevokeReasonSignOut)
	return nil
}

// openSession creates and persists a session, recording the session metric.
// Shared by AuthService and OAuthService.
func openSession(
	ctx context.Context,
	sessionRepo repo.ISessionRepository,
	sessionMetrics *metrics.SessionMetrics,
	userID uuid.UUID,
	ip, userAgent string,
) (*models.Session, error) {
	session := &models.Session{
		UserID:    userID,
		IP:        ip,
		UserAgent: userAgent,
	}
	if err := sessionRepo.CreateSession(ctx, session); err != nil {
		return nil, Internal("Failed to create session", err)
	}
	sessionMetrics.IncCreated()
	return session, nil
}

func authOutcomeFromError(err error) metrics.AuthOutcome {
	if err == nil {
		return metrics.AuthOutcomeSuccess
	}
	var se *Error
	if errors.As(err, &se) {
		switch se.Code {
		case "bad_request":
			return metrics.AuthOutcomeInvalidInput
		case "unauthorized":
			return metrics.AuthOutcomeInvalidCreds
		case "conflict":
			return metrics.AuthOutcomeUserAlreadyExist
		}
	}
	return metrics.AuthOutcomeInternalError
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

func generateCode() (string, error) {
	max := big.NewInt(1_000_000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
