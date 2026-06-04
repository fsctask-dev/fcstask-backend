package service

import (
	"context"
	"errors"
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
	"fcstask-backend/internal/oauth"
)

// OAuthService owns the OAuth provider flows: exchanging a provider payload for
// a session or a registration token, completing an OAuth-initiated signup, and
// linking/unlinking identities on an existing account.
type OAuthService struct {
	userRepo              repo.IUserRepo
	sessionRepo           repo.ISessionRepository
	emailRegistrationRepo repo.IEmailRegistrationRepo
	oauthRegistrationRepo repo.IRegistrationSessionRepo
	oauthRepo             repo.IOAuthIdentityRepo

	registry *oauth.Registry
	mailer   mailer.Mailer

	oauthConfig             config.OAuthConfig
	emailRegistrationConfig config.EmailRegistrationConfig

	sessionMetrics *metrics.SessionMetrics
	now            func() time.Time
}

func NewOAuthService(
	userRepo repo.IUserRepo,
	sessionRepo repo.ISessionRepository,
	emailRegistrationRepo repo.IEmailRegistrationRepo,
	oauthRegistrationRepo repo.IRegistrationSessionRepo,
	oauthRepo repo.IOAuthIdentityRepo,
	registry *oauth.Registry,
	m mailer.Mailer,
	oauthConfig config.OAuthConfig,
	emailRegistrationConfig config.EmailRegistrationConfig,
) *OAuthService {
	return &OAuthService{
		userRepo:                userRepo,
		sessionRepo:             sessionRepo,
		emailRegistrationRepo:   emailRegistrationRepo,
		oauthRegistrationRepo:   oauthRegistrationRepo,
		oauthRepo:               oauthRepo,
		registry:                registry,
		mailer:                  m,
		oauthConfig:             oauthConfig,
		emailRegistrationConfig: emailRegistrationConfig,
		now:                     time.Now,
	}
}

func (s *OAuthService) WithMetrics(session *metrics.SessionMetrics) *OAuthService {
	s.sessionMetrics = session
	return s
}

type OauthExchangeInput struct {
	ProviderName string
	Payload      oauth.ExchangePayload
	IP           string
	UserAgent    string
}

type oauthSuggestedProfile struct {
	Email     *string `json:"email,omitempty"`
	Username  *string `json:"username,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

type OauthRegistrationResult struct {
	Provider         string
	Token            uuid.UUID
	SuggestedProfile oauthSuggestedProfile
}

type OauthExchangeResult struct {
	RegistrationRequired bool

	Auth         *AuthResult
	Registration *OauthRegistrationResult
}

type OauthCompleteInput struct {
	RegistrationToken uuid.UUID
	Email             string
	Username          string
	Password          string
	FirstName         *string
	LastName          *string
}

func (s *OAuthService) OauthExchange(ctx context.Context, input OauthExchangeInput) (result *OauthExchangeResult, err error) {
	provider, ok := s.registry.Get(input.ProviderName)
	if !ok || !provider.Enabled() {
		return nil, NotFound("Oauth provider not found or disabled")
	}

	profile, err := provider.Exchange(ctx, input.Payload)
	if err != nil {
		switch {
		case errors.Is(err, oauth.ErrInvalidPayload),
			errors.Is(err, oauth.ErrSignatureMismatch),
			errors.Is(err, oauth.ErrPayloadExpired):
			return nil, BadRequest("Invalid oauth payload")
		case errors.Is(err, oauth.ErrProviderDisabled):
			return nil, NotFound("Oauth provider disabled")
		default:
			return nil, Internal("Oauth exchange failed", err)
		}
	}

	identity, err := s.oauthRepo.GetByProviderUID(ctx, provider.Name(), profile.ProviderUID)
	switch {
	case err == nil:
		applyProfileToIdentity(identity, profile)
		if err := s.oauthRepo.Update(ctx, identity); err != nil {
			return nil, Internal("Failed to update oauth identity", err)
		}
		return s.signInExisting(ctx, input, identity.UserID)

	case errors.Is(err, gorm.ErrRecordNotFound):
		// No identity yet → create a transient registration session.
		regSession := &models.RegistrationSession{
			Provider:    provider.Name(),
			ProviderUID: profile.ProviderUID,
			ExpiresAt:   s.now().Add(s.oauthConfig.RegistractionTTL),
		}
		applyProfileToRegistration(regSession, profile)
		if err := s.oauthRegistrationRepo.Create(ctx, regSession); err != nil {
			return nil, Internal("Failed to create registration session", err)
		}

		suggested := oauthSuggestedProfile{}
		if profile.Email != "" {
			e := profile.Email
			suggested.Email = &e
		}
		if profile.Username != "" {
			u := profile.Username
			suggested.Username = &u
		}
		if profile.FirstName != "" {
			f := profile.FirstName
			suggested.FirstName = &f
		}
		if profile.LastName != "" {
			l := profile.LastName
			suggested.LastName = &l
		}
		result = &OauthExchangeResult{
			RegistrationRequired: true,
			Registration: &OauthRegistrationResult{
				SuggestedProfile: suggested,
				Token:            regSession.ID,
				Provider:         provider.Name(),
			},
		}
		return result, nil
	default:
		return nil, Internal("Failed to look up OAuth identity", err)
	}
}

func (s *OAuthService) signInExisting(ctx context.Context, input OauthExchangeInput, userID uuid.UUID) (result *OauthExchangeResult, err error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, Internal("Failed to load user", err)
	}

	session, err := openSession(ctx, s.sessionRepo, s.sessionMetrics, user.ID, input.IP, input.UserAgent)
	if err != nil {
		return nil, err
	}

	result = &OauthExchangeResult{
		RegistrationRequired: false,
		Auth: &AuthResult{
			Session: session,
			User:    user,
		},
	}
	return result, nil
}

func (s *OAuthService) OauthComplete(ctx context.Context, input OauthCompleteInput) (result *model.EmailRegistration, err error) {
	if input.Username == "" || input.Email == "" || input.RegistrationToken == uuid.Nil {
		return nil, BadRequest("registration_token, username and email are required")
	}

	regSession, err := s.oauthRegistrationRepo.GetByID(ctx, input.RegistrationToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NotFound("Registration token not found")
		}
		return nil, Internal("Failed to load registration session", err)
	}
	if s.now().UTC().After(regSession.ExpiresAt) {
		if err := s.oauthRegistrationRepo.Delete(ctx, regSession.ID); err != nil {
			return nil, Internal("Failed to delete expired registration", err)
		}

		return nil, NotFound("Registration token has expired")
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

	_ = s.oauthRegistrationRepo.Delete(ctx, regSession.ID)

	identity := &models.OAuthIdentity{
		EmailRegistrationID: result.ID,
		Provider:            regSession.Provider,
		ProviderUID:         regSession.ProviderUID,
		Username:            regSession.Username,
		AccessToken:         regSession.AccessToken,
		RefreshToken:        regSession.RefreshToken,
		RawProfile:          regSession.RawProfile,
	}
	if err := s.oauthRepo.Create(ctx, identity); err != nil {
		return nil, Internal("Failed to create OAuth identity", err)
	}
	return result, nil
}

func (s *OAuthService) OauthAddExchange(ctx context.Context, user *model.User, input OauthExchangeInput) (err error) {
	provider, ok := s.registry.Get(input.ProviderName)
	if !ok || !provider.Enabled() {
		return NotFound("Oauth provider not found or disabled")
	}

	profile, err := provider.Exchange(ctx, input.Payload)
	if err != nil {
		switch {
		case errors.Is(err, oauth.ErrInvalidPayload),
			errors.Is(err, oauth.ErrSignatureMismatch),
			errors.Is(err, oauth.ErrPayloadExpired):
			return BadRequest("Invalid oauth payload")
		case errors.Is(err, oauth.ErrProviderDisabled):
			return NotFound("Oauth provider disabled")
		default:
			return Internal("Oauth exchange failed", err)
		}
	}

	identity, err := s.oauthRepo.GetProviderForUserID(ctx, user.ID, provider.Name())
	switch {
	case err == nil:
		applyProfileToIdentity(identity, profile)
		if err := s.oauthRepo.Update(ctx, identity); err != nil {
			return Internal("Failed to update oauth identity", err)
		}

	case errors.Is(err, gorm.ErrRecordNotFound):
		// No identity yet
		identity := &models.OAuthIdentity{
			UserID:       user.ID,
			Provider:     provider.Name(),
			ProviderUID:  profile.ProviderUID,
			AccessToken:  profile.AccessToken,
			RefreshToken: profile.RefreshToken,
			RawProfile:   profile.Raw,
		}
		applyProfileToIdentity(identity, profile)

		if err := s.oauthRepo.Create(ctx, identity); err != nil {
			return Internal("Failed to create oauth identity", err)
		}
	default:
		return Internal("Failed to look up OAuth identity", err)
	}
	return nil
}

func (s *OAuthService) OauthUnlink(ctx context.Context, user *model.User, providerName string) (err error) {
	provider, ok := s.registry.Get(providerName)
	if !ok || !provider.Enabled() {
		return NotFound("Oauth provider not found or disabled")
	}

	identity, err := s.oauthRepo.GetProviderForUserID(ctx, user.ID, provider.Name())
	switch {
	case err == nil:
		if err := s.oauthRepo.Delete(ctx, identity.ID); err != nil {
			return Internal("Failed to create oauth identity", err)
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		return NotFound("Identity not found")
	default:
		return Internal("Failed to look up OAuth identity", err)
	}
	return nil
}

// applyProfileToIdentity copies all provider-derived fields from the freshly
// exchanged profile onto the stored identity. For Telegram this persists the
// telegram id (ProviderUID) and handle (Username) on the identity.
func applyProfileToIdentity(identity *models.OAuthIdentity, profile *oauth.Profile) {
	identity.ProviderUID = profile.ProviderUID
	identity.Username = profile.Username
	identity.AccessToken = profile.AccessToken
	identity.RefreshToken = profile.RefreshToken
	identity.ExpiresAt = profile.ExpiresAt
	identity.RawProfile = profile.Raw
}

func applyProfileToRegistration(reg *models.RegistrationSession, profile *oauth.Profile) {
	reg.Email = profile.Email
	reg.Username = profile.Username
	reg.FirstName = profile.FirstName
	reg.LastName = profile.LastName
	reg.AccessToken = profile.AccessToken
	reg.RefreshToken = profile.RefreshToken
	reg.TokenExpiresAt = profile.ExpiresAt
	reg.RawProfile = profile.Raw
}
