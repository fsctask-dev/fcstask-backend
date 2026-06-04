package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"fcstask-backend/internal/config"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/oauth"
)

// fakeProvider is a configurable oauth.Provider for tests.
type fakeProvider struct {
	name    string
	enabled bool
	profile *oauth.Profile
	err     error
}

func (p *fakeProvider) Name() string  { return p.name }
func (p *fakeProvider) Enabled() bool { return p.enabled }
func (p *fakeProvider) Exchange(ctx context.Context, payload oauth.ExchangePayload) (*oauth.Profile, error) {
	return p.profile, p.err
}

// oauthIdentityRepoStub is a configurable IOAuthIdentityRepo.
type oauthIdentityRepoStub struct {
	identity *models.OAuthIdentity // returned by lookups when set, else NotFound
	created  *models.OAuthIdentity
	updated  bool
	deleted  bool
}

func (r *oauthIdentityRepoStub) Create(ctx context.Context, i *models.OAuthIdentity) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	r.created = i
	return nil
}
func (r *oauthIdentityRepoStub) Update(ctx context.Context, i *models.OAuthIdentity) error {
	r.updated = true
	r.identity = i
	return nil
}
func (r *oauthIdentityRepoStub) GetByID(ctx context.Context, id uuid.UUID) (*models.OAuthIdentity, error) {
	if r.identity == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.identity, nil
}
func (r *oauthIdentityRepoStub) GetByProviderUID(ctx context.Context, provider, uid string) (*models.OAuthIdentity, error) {
	if r.identity == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.identity, nil
}
func (r *oauthIdentityRepoStub) GetProviderForUserID(ctx context.Context, userID uuid.UUID, provider string) (*models.OAuthIdentity, error) {
	if r.identity == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.identity, nil
}
func (r *oauthIdentityRepoStub) GetByEmailRegistrationID(ctx context.Context, id uuid.UUID) (*models.OAuthIdentity, error) {
	return nil, gorm.ErrRecordNotFound
}
func (r *oauthIdentityRepoStub) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.OAuthIdentity, error) {
	return nil, nil
}
func (r *oauthIdentityRepoStub) Delete(ctx context.Context, id uuid.UUID) error {
	r.deleted = true
	return nil
}

type registrationSessionRepoStub struct {
	session *models.RegistrationSession
	created bool
	deleted bool
}

func (r *registrationSessionRepoStub) Create(ctx context.Context, s *models.RegistrationSession) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	r.session = s
	r.created = true
	return nil
}
func (r *registrationSessionRepoStub) GetByID(ctx context.Context, id uuid.UUID) (*models.RegistrationSession, error) {
	if r.session == nil || r.session.ID != id {
		return nil, gorm.ErrRecordNotFound
	}
	return r.session, nil
}
func (r *registrationSessionRepoStub) Delete(ctx context.Context, id uuid.UUID) error {
	r.deleted = true
	return nil
}
func (r *registrationSessionRepoStub) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

func newOAuthServiceForTest(
	userRepo *authUserRepo,
	sessionRepo *authSessionRepo,
	emailRepo *authEmailRegRepo,
	regRepo *registrationSessionRepoStub,
	oauthRepo *oauthIdentityRepoStub,
	provider oauth.Provider,
) *OAuthService {
	return NewOAuthService(
		userRepo, sessionRepo, emailRepo, regRepo, oauthRepo,
		oauth.NewRegistry(provider), &stubMailer{},
		config.OAuthConfig{RegistractionTTL: 15 * time.Minute},
		config.EmailRegistrationConfig{TTL: 15 * time.Minute},
	)
}

func TestOauthExchange_RegistrationRequired(t *testing.T) {
	provider := &fakeProvider{name: "gitlab", enabled: true, profile: &oauth.Profile{
		ProviderUID: "123",
		Email:       "g@example.com",
		Username:    "ghandle",
	}}
	regRepo := &registrationSessionRepoStub{}
	oauthRepo := &oauthIdentityRepoStub{} // no existing identity → NotFound
	svc := newOAuthServiceForTest(&authUserRepo{}, &authSessionRepo{}, &authEmailRegRepo{}, regRepo, oauthRepo, provider)

	result, err := svc.OauthExchange(context.Background(), OauthExchangeInput{ProviderName: "gitlab"})

	assert.NoError(t, err)
	assert.True(t, result.RegistrationRequired)
	assert.Nil(t, result.Auth)
	assert.True(t, regRepo.created)
	assert.Equal(t, regRepo.session.ID, result.Registration.Token)
	assert.Equal(t, "gitlab", result.Registration.Provider)
	assert.Equal(t, "g@example.com", *result.Registration.SuggestedProfile.Email)
	assert.Equal(t, "ghandle", *result.Registration.SuggestedProfile.Username)
}

func TestOauthExchange_ExistingIdentitySignsIn(t *testing.T) {
	userID := uuid.New()
	provider := &fakeProvider{name: "gitlab", enabled: true, profile: &oauth.Profile{ProviderUID: "123"}}
	oauthRepo := &oauthIdentityRepoStub{identity: &models.OAuthIdentity{ID: uuid.New(), UserID: userID, Provider: "gitlab"}}
	userRepo := &authUserRepo{user: &models.User{ID: userID, Email: "g@example.com", Username: "ghandle"}}
	sessionRepo := &authSessionRepo{}
	svc := newOAuthServiceForTest(userRepo, sessionRepo, &authEmailRegRepo{}, &registrationSessionRepoStub{}, oauthRepo, provider)

	result, err := svc.OauthExchange(context.Background(), OauthExchangeInput{ProviderName: "gitlab"})

	assert.NoError(t, err)
	assert.False(t, result.RegistrationRequired)
	assert.NotNil(t, result.Auth)
	assert.Equal(t, userID, result.Auth.Session.UserID)
	assert.True(t, oauthRepo.updated, "identity tokens refreshed on sign-in")
}

func TestOauthExchange_ProviderDisabled(t *testing.T) {
	provider := &fakeProvider{name: "gitlab", enabled: false}
	svc := newOAuthServiceForTest(&authUserRepo{}, &authSessionRepo{}, &authEmailRegRepo{}, &registrationSessionRepoStub{}, &oauthIdentityRepoStub{}, provider)

	result, err := svc.OauthExchange(context.Background(), OauthExchangeInput{ProviderName: "gitlab"})

	assert.Nil(t, result)
	assertServiceCode(t, err, "not_found")
}

func TestOauthComplete_Success(t *testing.T) {
	regRepo := &registrationSessionRepoStub{}
	token := uuid.New()
	regRepo.session = &models.RegistrationSession{
		ID:          token,
		Provider:    "gitlab",
		ProviderUID: "123",
		Username:    "ghandle",
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}
	emailRepo := &authEmailRegRepo{}
	oauthRepo := &oauthIdentityRepoStub{}
	svc := newOAuthServiceForTest(&authUserRepo{}, &authSessionRepo{}, emailRepo, regRepo, oauthRepo, &fakeProvider{name: "gitlab", enabled: true})

	reg, err := svc.OauthComplete(context.Background(), OauthCompleteInput{
		RegistrationToken: token,
		Email:             "g@example.com",
		Username:          "ghandle",
	})

	assert.NoError(t, err)
	assert.NotNil(t, reg)
	// A pending email registration and a linked identity are created.
	assert.NotNil(t, emailRepo.reg)
	assert.NotNil(t, oauthRepo.created)
	assert.Equal(t, reg.ID, oauthRepo.created.EmailRegistrationID)
	assert.Equal(t, "ghandle", oauthRepo.created.Username)
}

func TestOauthComplete_ExpiredToken(t *testing.T) {
	token := uuid.New()
	regRepo := &registrationSessionRepoStub{session: &models.RegistrationSession{
		ID:        token,
		ExpiresAt: time.Now().Add(-time.Minute),
	}}
	svc := newOAuthServiceForTest(&authUserRepo{}, &authSessionRepo{}, &authEmailRegRepo{}, regRepo, &oauthIdentityRepoStub{}, &fakeProvider{name: "gitlab", enabled: true})

	reg, err := svc.OauthComplete(context.Background(), OauthCompleteInput{
		RegistrationToken: token,
		Email:             "g@example.com",
		Username:          "ghandle",
	})

	assert.Nil(t, reg)
	assertServiceCode(t, err, "not_found")
	assert.True(t, regRepo.deleted, "expired registration session is cleaned up")
}

func TestOauthUnlink_Success(t *testing.T) {
	oauthRepo := &oauthIdentityRepoStub{identity: &models.OAuthIdentity{ID: uuid.New(), Provider: "gitlab"}}
	svc := newOAuthServiceForTest(&authUserRepo{}, &authSessionRepo{}, &authEmailRegRepo{}, &registrationSessionRepoStub{}, oauthRepo, &fakeProvider{name: "gitlab", enabled: true})

	err := svc.OauthUnlink(context.Background(), &models.User{ID: uuid.New()}, "gitlab")

	assert.NoError(t, err)
	assert.True(t, oauthRepo.deleted)
}

func TestOauthUnlink_NotLinked(t *testing.T) {
	svc := newOAuthServiceForTest(&authUserRepo{}, &authSessionRepo{}, &authEmailRegRepo{}, &registrationSessionRepoStub{}, &oauthIdentityRepoStub{}, &fakeProvider{name: "gitlab", enabled: true})

	err := svc.OauthUnlink(context.Background(), &models.User{ID: uuid.New()}, "gitlab")

	assertServiceCode(t, err, "not_found")
}
