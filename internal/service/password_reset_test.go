package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"fcstask-backend/internal/config"
	models "fcstask-backend/internal/db/model"
)

type authPasswordResetRepo struct {
	pr      *models.PasswordReset
	created bool
	deleted bool
}

func (r *authPasswordResetRepo) Create(ctx context.Context, pr *models.PasswordReset) error {
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	r.pr = pr
	r.created = true
	return nil
}
func (r *authPasswordResetRepo) Update(ctx context.Context, pr *models.PasswordReset) error {
	r.pr = pr
	return nil
}
func (r *authPasswordResetRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.PasswordReset, error) {
	if r.pr == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.pr, nil
}
func (r *authPasswordResetRepo) GetByUserEmail(ctx context.Context, email string) (*models.PasswordReset, error) {
	if r.pr == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.pr, nil
}
func (r *authPasswordResetRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.deleted = true
	r.pr = nil
	return nil
}
func (r *authPasswordResetRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (r *authPasswordResetRepo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

func newPasswordResetServiceForTest(userRepo *authUserRepo, prRepo *authPasswordResetRepo, mail *stubMailer) *PasswordResetService {
	return NewPasswordResetService(userRepo, prRepo, mail, config.EmailRegistrationConfig{TTL: 15 * time.Minute})
}

func TestPasswordResetRequest_Success(t *testing.T) {
	userRepo := &authUserRepo{user: &models.User{ID: uuid.New(), Email: "u@example.com", Username: "u"}}
	prRepo := &authPasswordResetRepo{}
	mail := &stubMailer{}
	svc := newPasswordResetServiceForTest(userRepo, prRepo, mail)

	pr, err := svc.PasswordResetRequest(context.Background(), "u@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, userRepo.user.ID, pr.UserID)
	assert.NotEmpty(t, pr.CodeHash)
	assert.True(t, pr.ExpiresAt.After(time.Now()))
	assert.Equal(t, 1, mail.sent)
	assert.True(t, prRepo.created)
}

func TestPasswordResetRequest_UserNotFound(t *testing.T) {
	svc := newPasswordResetServiceForTest(&authUserRepo{}, &authPasswordResetRepo{}, &stubMailer{})

	pr, err := svc.PasswordResetRequest(context.Background(), "missing@example.com")

	assert.Nil(t, pr)
	assertServiceCode(t, err, "not_found")
}

func TestPasswordResetConfirm_Success(t *testing.T) {
	const code = "123456"
	codeHash, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	userID := uuid.New()
	userRepo := &authUserRepo{user: &models.User{ID: userID, Email: "u@example.com"}}
	prRepo := &authPasswordResetRepo{pr: &models.PasswordReset{
		ID:        uuid.New(),
		UserID:    userID,
		CodeHash:  string(codeHash),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}}
	svc := newPasswordResetServiceForTest(userRepo, prRepo, &stubMailer{})

	err := svc.PasswordResetConfirm(context.Background(), PasswordResetConfirmInput{
		Email:       "u@example.com",
		Code:        code,
		Password:    "newsecret",
		MaxAttempts: 5,
	})

	assert.NoError(t, err)
	// Password updated to a hash of the new password, reset consumed.
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(userRepo.user.PasswordHash), []byte("newsecret")))
	assert.True(t, prRepo.deleted)
}

func TestPasswordResetConfirm_WrongCode(t *testing.T) {
	codeHash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	prRepo := &authPasswordResetRepo{pr: &models.PasswordReset{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		CodeHash:  string(codeHash),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}}
	svc := newPasswordResetServiceForTest(&authUserRepo{}, prRepo, &stubMailer{})

	err := svc.PasswordResetConfirm(context.Background(), PasswordResetConfirmInput{
		Email:       "u@example.com",
		Code:        "000000",
		Password:    "newsecret",
		MaxAttempts: 5,
	})

	assertServiceCode(t, err, "unauthorized")
	assert.Equal(t, 1, prRepo.pr.Attempts)
}

func TestPasswordResetConfirm_Expired(t *testing.T) {
	codeHash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	prRepo := &authPasswordResetRepo{pr: &models.PasswordReset{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		CodeHash:  string(codeHash),
		ExpiresAt: time.Now().Add(-time.Minute),
	}}
	svc := newPasswordResetServiceForTest(&authUserRepo{}, prRepo, &stubMailer{})

	err := svc.PasswordResetConfirm(context.Background(), PasswordResetConfirmInput{
		Email:       "u@example.com",
		Code:        "123456",
		Password:    "newsecret",
		MaxAttempts: 5,
	})

	assertServiceCode(t, err, "unauthorized")
}
