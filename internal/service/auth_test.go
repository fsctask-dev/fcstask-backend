package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/openframebox/gomail"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"fcstask-backend/internal/config"
	models "fcstask-backend/internal/db/model"
)

type authUserRepo struct {
	user          *models.User
	emailTaken    bool
	usernameTaken bool
	createErr     error
	findErr       error
}

func (r *authUserRepo) CreateUser(ctx context.Context, user *models.User) error {
	if r.createErr != nil {
		return r.createErr
	}
	user.ID = uuid.New()
	r.user = user
	return nil
}

func (r *authUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if r.user == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *authUserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	if r.user == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *authUserRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	if r.user == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *authUserRepo) GetUserByUserID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *authUserRepo) GetUserByTgUID(ctx context.Context, tgUID int64) (*models.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *authUserRepo) UpdateUser(ctx context.Context, user *models.User) error {
	r.user = user
	return nil
}

func (r *authUserRepo) DeleteUser(ctx context.Context, id uuid.UUID) error {
	r.user = nil
	return nil
}

func (r *authUserRepo) GetUsersWithSessions(ctx context.Context, limit, offset int) ([]models.User, error) {
	return nil, nil
}

func (r *authUserRepo) CountUsersWithSessions(ctx context.Context) (int64, error) {
	return 0, nil
}

func (r *authUserRepo) ExistsUserByEmail(ctx context.Context, email string) (bool, error) {
	return r.emailTaken, nil
}

func (r *authUserRepo) ExistsUserByUsername(ctx context.Context, username string) (bool, error) {
	return r.usernameTaken, nil
}

func (r *authUserRepo) CountUsers(ctx context.Context) (int64, error) {
	return 0, nil
}

type authSessionRepo struct {
	session   *models.Session
	createErr error
	deleteErr error
}

func (r *authSessionRepo) CreateSession(ctx context.Context, session *models.Session) error {
	if r.createErr != nil {
		return r.createErr
	}
	session.ID = uuid.New()
	r.session = session
	return nil
}

func (r *authSessionRepo) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	if r.session == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.session, nil
}

func (r *authSessionRepo) GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	return nil, nil
}

func (r *authSessionRepo) GetSessionsWithUser(ctx context.Context, limit, offset int) ([]models.Session, error) {
	return nil, nil
}

func (r *authSessionRepo) CountSessions(ctx context.Context) (int64, error) {
	return 0, nil
}

func (r *authSessionRepo) TouchSessionAccessedAt(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (r *authSessionRepo) DeleteSession(ctx context.Context, id uuid.UUID) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	r.session = nil
	return nil
}

func (r *authSessionRepo) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (r *authSessionRepo) CleanOutdatedSessions(ctx context.Context, ttl time.Duration) (int64, error) {
	return 0, nil
}

type authEmailRegRepo struct {
	reg            *models.EmailRegistration
	createErr      error
	getErr         error
	deleted        bool
	deletedByEmail bool
}

func (r *authEmailRegRepo) Create(ctx context.Context, reg *models.EmailRegistration) error {
	if r.createErr != nil {
		return r.createErr
	}
	if reg.ID == uuid.Nil {
		reg.ID = uuid.New()
	}
	r.reg = reg
	return nil
}

func (r *authEmailRegRepo) Update(ctx context.Context, reg *models.EmailRegistration) error {
	r.reg = reg
	return nil
}

func (r *authEmailRegRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.EmailRegistration, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if r.reg == nil || r.reg.ID != id {
		return nil, gorm.ErrRecordNotFound
	}
	return r.reg, nil
}

func (r *authEmailRegRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.deleted = true
	r.reg = nil
	return nil
}

func (r *authEmailRegRepo) DeleteByEmail(ctx context.Context, email string) error {
	r.deletedByEmail = true
	return nil
}

func (r *authEmailRegRepo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

// stubMailer satisfies mailer.Mailer and records how many messages were sent.
type stubMailer struct {
	sent int
	err  error
}

func (m *stubMailer) Send(ctx context.Context, to gomail.Address, subject, body string) error {
	m.sent++
	return m.err
}

// authOAuthRepo is a no-op IOAuthIdentityRepo. GetByEmailRegistrationID returns
// a nil error (mirroring the real Find-based repo) so SignUpVerify treats it as
// "no identity to relink" and proceeds.
type authOAuthRepo struct{}

func (authOAuthRepo) Create(ctx context.Context, i *models.OAuthIdentity) error { return nil }
func (authOAuthRepo) Update(ctx context.Context, i *models.OAuthIdentity) error { return nil }
func (authOAuthRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.OAuthIdentity, error) {
	return nil, gorm.ErrRecordNotFound
}
func (authOAuthRepo) GetByProviderUID(ctx context.Context, provider, uid string) (*models.OAuthIdentity, error) {
	return nil, gorm.ErrRecordNotFound
}
func (authOAuthRepo) GetProviderForUserID(ctx context.Context, userID uuid.UUID, provider string) (*models.OAuthIdentity, error) {
	return nil, gorm.ErrRecordNotFound
}
func (authOAuthRepo) GetByEmailRegistrationID(ctx context.Context, id uuid.UUID) (*models.OAuthIdentity, error) {
	return &models.OAuthIdentity{}, nil
}
func (authOAuthRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.OAuthIdentity, error) {
	return nil, nil
}
func (authOAuthRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

func newAuthServiceForTest(userRepo *authUserRepo, sessionRepo *authSessionRepo, emailRepo *authEmailRegRepo, mail *stubMailer) *AuthService {
	return NewAuthService(
		userRepo, sessionRepo, emailRepo, authOAuthRepo{}, mail,
		config.EmailRegistrationConfig{TTL: 15 * time.Minute},
	)
}

func assertServiceCode(t *testing.T, err error, code string) {
	t.Helper()
	var serviceErr *Error
	assert.True(t, errors.As(err, &serviceErr), "expected *service.Error, got %v", err)
	if serviceErr != nil {
		assert.Equal(t, code, serviceErr.Code)
	}
}

func TestAuthService_SignUpBeginSuccess(t *testing.T) {
	userRepo := &authUserRepo{}
	emailRepo := &authEmailRegRepo{}
	mail := &stubMailer{}
	svc := newAuthServiceForTest(userRepo, nil, emailRepo, mail)

	reg, err := svc.SignUp(context.Background(), SignUpInput{
		Email:    "new@example.com",
		Username: "newuser",
		Password: "secret",
	})

	assert.NoError(t, err)
	assert.NotNil(t, reg)
	assert.Equal(t, "new@example.com", reg.Email)
	assert.Equal(t, "newuser", reg.Username)
	// Password and code must be stored hashed, never in the clear.
	assert.NotEmpty(t, reg.PasswordHash)
	assert.NotEqual(t, "secret", reg.PasswordHash)
	assert.NotEmpty(t, reg.CodeHash)
	assert.True(t, reg.ExpiresAt.After(time.Now()))
	// A verification email is sent and the pending row is persisted...
	assert.Equal(t, 1, mail.sent)
	assert.True(t, emailRepo.deletedByEmail, "previous pending registration should be cleared")
	assert.NotNil(t, emailRepo.reg)
	// ...but the user is NOT created until verification.
	assert.Nil(t, userRepo.user)
}

func TestAuthService_SignUpMissingFields(t *testing.T) {
	svc := newAuthServiceForTest(&authUserRepo{}, nil, &authEmailRegRepo{}, &stubMailer{})

	reg, err := svc.SignUp(context.Background(), SignUpInput{
		Email:    "new@example.com",
		Username: "",
		Password: "secret",
	})

	assert.Nil(t, reg)
	assertServiceCode(t, err, "bad_request")
}

func TestAuthService_SignUpEmailConflict(t *testing.T) {
	emailRepo := &authEmailRegRepo{}
	svc := newAuthServiceForTest(&authUserRepo{emailTaken: true}, nil, emailRepo, &stubMailer{})

	reg, err := svc.SignUp(context.Background(), SignUpInput{
		Email:    "taken@example.com",
		Username: "newuser",
		Password: "secret",
	})

	assert.Nil(t, reg)
	assertServiceCode(t, err, "conflict")
	assert.Nil(t, emailRepo.reg, "no registration should be created on conflict")
}

func TestAuthService_SignUpUsernameConflict(t *testing.T) {
	svc := newAuthServiceForTest(&authUserRepo{usernameTaken: true}, nil, &authEmailRegRepo{}, &stubMailer{})

	reg, err := svc.SignUp(context.Background(), SignUpInput{
		Email:    "new@example.com",
		Username: "taken",
		Password: "secret",
	})

	assert.Nil(t, reg)
	assertServiceCode(t, err, "conflict")
}

func TestAuthService_SignUpMailerError(t *testing.T) {
	emailRepo := &authEmailRegRepo{}
	mail := &stubMailer{err: errors.New("smtp down")}
	svc := newAuthServiceForTest(&authUserRepo{}, nil, emailRepo, mail)

	reg, err := svc.SignUp(context.Background(), SignUpInput{
		Email:    "new@example.com",
		Username: "newuser",
		Password: "secret",
	})

	assert.Nil(t, reg)
	assertServiceCode(t, err, "internal_error")
	assert.Nil(t, emailRepo.reg, "registration must not be persisted if the email fails")
}

func TestAuthService_SignUpVerifySuccess(t *testing.T) {
	const code = "123456"
	codeHash, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	pwHash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	regID := uuid.New()
	emailRepo := &authEmailRegRepo{reg: &models.EmailRegistration{
		ID:           regID,
		Email:        "new@example.com",
		Username:     "newuser",
		PasswordHash: string(pwHash),
		CodeHash:     string(codeHash),
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}}
	userRepo := &authUserRepo{}
	sessionRepo := &authSessionRepo{}
	svc := newAuthServiceForTest(userRepo, sessionRepo, emailRepo, &stubMailer{})

	result, err := svc.SignUpVerify(context.Background(), SignUpVerifyInput{
		Token:       regID,
		Code:        code,
		MaxAttempts: 5,
		IP:          "127.0.0.1",
		UserAgent:   "test-agent",
	})

	assert.NoError(t, err)
	assert.Equal(t, "new@example.com", result.User.Email)
	assert.Equal(t, "newuser", result.User.Username)
	assert.Equal(t, string(pwHash), result.User.PasswordHash, "hashed password carries over unchanged")
	assert.NotEqual(t, uuid.Nil, result.Session.ID)
	assert.Equal(t, result.User.ID, result.Session.UserID)
	assert.NotNil(t, userRepo.user, "user is created on verify")
	assert.True(t, emailRepo.deleted, "registration is consumed on success")
}

func TestAuthService_SignUpVerifyTokenNotFound(t *testing.T) {
	svc := newAuthServiceForTest(&authUserRepo{}, &authSessionRepo{}, &authEmailRegRepo{}, &stubMailer{})

	result, err := svc.SignUpVerify(context.Background(), SignUpVerifyInput{
		Token:       uuid.New(),
		Code:        "123456",
		MaxAttempts: 5,
	})

	assert.Nil(t, result)
	assertServiceCode(t, err, "not_found")
}

func TestAuthService_SignUpVerifyExpired(t *testing.T) {
	const code = "123456"
	codeHash, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	regID := uuid.New()
	emailRepo := &authEmailRegRepo{reg: &models.EmailRegistration{
		ID:        regID,
		CodeHash:  string(codeHash),
		ExpiresAt: time.Now().Add(-time.Minute),
	}}
	svc := newAuthServiceForTest(&authUserRepo{}, &authSessionRepo{}, emailRepo, &stubMailer{})

	result, err := svc.SignUpVerify(context.Background(), SignUpVerifyInput{
		Token:       regID,
		Code:        code,
		MaxAttempts: 5,
	})

	assert.Nil(t, result)
	assertServiceCode(t, err, "unauthorized")
	assert.NotNil(t, emailRepo.reg, "expired registration is left untouched")
}

func TestAuthService_SignUpVerifyWrongCode(t *testing.T) {
	codeHash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	regID := uuid.New()
	emailRepo := &authEmailRegRepo{reg: &models.EmailRegistration{
		ID:        regID,
		CodeHash:  string(codeHash),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}}
	userRepo := &authUserRepo{}
	svc := newAuthServiceForTest(userRepo, &authSessionRepo{}, emailRepo, &stubMailer{})

	result, err := svc.SignUpVerify(context.Background(), SignUpVerifyInput{
		Token:       regID,
		Code:        "000000",
		MaxAttempts: 5,
	})

	assert.Nil(t, result)
	assertServiceCode(t, err, "unauthorized")
	assert.Equal(t, 1, emailRepo.reg.Attempts, "a bad code increments the attempt counter")
	assert.Nil(t, userRepo.user, "no user created on a bad code")
}

func TestAuthService_SignUpVerifyAttemptsExceeded(t *testing.T) {
	codeHash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	regID := uuid.New()
	emailRepo := &authEmailRegRepo{reg: &models.EmailRegistration{
		ID:        regID,
		CodeHash:  string(codeHash),
		Attempts:  6,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}}
	svc := newAuthServiceForTest(&authUserRepo{}, &authSessionRepo{}, emailRepo, &stubMailer{})

	result, err := svc.SignUpVerify(context.Background(), SignUpVerifyInput{
		Token:       regID,
		Code:        "123456",
		MaxAttempts: 5,
	})

	assert.Nil(t, result)
	assertServiceCode(t, err, "unauthorized")
}

func TestAuthService_SignUpResendSuccess(t *testing.T) {
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("000000"), bcrypt.DefaultCost)
	regID := uuid.New()
	emailRepo := &authEmailRegRepo{reg: &models.EmailRegistration{
		ID:        regID,
		Email:     "new@example.com",
		Username:  "newuser",
		CodeHash:  string(oldHash),
		Attempts:  3,
		ExpiresAt: time.Now().Add(time.Minute),
	}}
	mail := &stubMailer{}
	svc := newAuthServiceForTest(&authUserRepo{}, nil, emailRepo, mail)

	reg, err := svc.SignUpResend(context.Background(), regID)

	assert.NoError(t, err)
	assert.NotNil(t, reg)
	assert.Equal(t, 0, reg.Attempts, "resend resets the attempt counter")
	assert.NotEqual(t, string(oldHash), reg.CodeHash, "a fresh code is issued")
	assert.Equal(t, 1, mail.sent)
	assert.True(t, reg.ExpiresAt.After(time.Now().Add(14*time.Minute)))
}

func TestAuthService_SignUpResendTokenNotFound(t *testing.T) {
	svc := newAuthServiceForTest(&authUserRepo{}, nil, &authEmailRegRepo{}, &stubMailer{})

	reg, err := svc.SignUpResend(context.Background(), uuid.New())

	assert.Nil(t, reg)
	assertServiceCode(t, err, "not_found")
}

func TestAuthService_SignUpResendExpired(t *testing.T) {
	regID := uuid.New()
	emailRepo := &authEmailRegRepo{reg: &models.EmailRegistration{
		ID:        regID,
		ExpiresAt: time.Now().Add(-time.Minute),
	}}
	mail := &stubMailer{}
	svc := newAuthServiceForTest(&authUserRepo{}, nil, emailRepo, mail)

	reg, err := svc.SignUpResend(context.Background(), regID)

	assert.Nil(t, reg)
	assertServiceCode(t, err, "unauthorized")
	assert.Equal(t, 0, mail.sent, "no email is sent for an expired registration")
}

func TestAuthService_SignInSuccess(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	userID := uuid.New()
	svc := NewAuthService(&authUserRepo{user: &models.User{
		ID:           userID,
		Email:        "user@example.com",
		Username:     "user",
		PasswordHash: string(hash),
	}}, &authSessionRepo{}, nil, nil, nil, config.EmailRegistrationConfig{})
	email := "user@example.com"

	result, err := svc.SignIn(context.Background(), SignInInput{
		Email:    &email,
		Password: "secret",
	})

	assert.NoError(t, err)
	assert.Equal(t, userID, result.Session.UserID)
}

func TestAuthService_SignInWrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	svc := NewAuthService(&authUserRepo{user: &models.User{
		ID:           uuid.New(),
		PasswordHash: string(hash),
	}}, &authSessionRepo{}, nil, nil, nil, config.EmailRegistrationConfig{})
	email := "user@example.com"

	result, err := svc.SignIn(context.Background(), SignInInput{
		Email:    &email,
		Password: "wrong",
	})

	assert.Nil(t, result)
	var serviceErr *Error
	assert.True(t, errors.As(err, &serviceErr))
	assert.Equal(t, "unauthorized", serviceErr.Code)
}

func TestAuthService_SignOutSuccess(t *testing.T) {
	sessionID := uuid.New()
	sessionRepo := &authSessionRepo{session: &models.Session{ID: sessionID}}
	svc := NewAuthService(nil, sessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})

	err := svc.SignOut(context.Background(), &models.Session{ID: sessionID})

	assert.NoError(t, err)
	assert.Nil(t, sessionRepo.session)
}
