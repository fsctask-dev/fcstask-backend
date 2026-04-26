package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"gorm.io/gorm"

	"fcstask-backend/internal/api"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
	"fcstask-backend/internal/oauth"
)

// registrationSessionTTL bounds how long a registration_token is valid.
// 30 minutes is enough for a human to fill out the signup form, short enough
// that a leaked token expires before it matters.
const registrationSessionTTL = 30 * time.Minute

// Response shapes are spelled out here instead of reusing api.* types because
// oapi-codegen v2 inlines schemas referenced across files (see schemas/*.yaml),
// which leaves us with anonymous nested structs that are awkward to construct.
// These structs match the JSON contract documented in the OpenAPI spec.

type oauthUserResponse struct {
	Id        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Username  string     `json:"username"`
	FirstName *string    `json:"first_name,omitempty"`
	LastName  *string    `json:"last_name,omitempty"`
	TgUid     *int64     `json:"tg_uid,omitempty"`
	UserId    uuid.UUID  `json:"user_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type oauthAuthResponse struct {
	SessionToken uuid.UUID         `json:"session_token"`
	User         oauthUserResponse `json:"user"`
}

type oauthSuggestedProfile struct {
	Email     *string `json:"email,omitempty"`
	Username  *string `json:"username,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

type oauthExchangeResponse struct {
	Status            string                 `json:"status"`
	Auth              *oauthAuthResponse     `json:"auth,omitempty"`
	RegistrationToken *uuid.UUID             `json:"registration_token,omitempty"`
	Suggested         *oauthSuggestedProfile `json:"suggested,omitempty"`
	Provider          *string                `json:"provider,omitempty"`
}

const (
	oauthStatusOK                   = "ok"
	oauthStatusRegistrationRequired = "registration_required"
)

type OAuthHandler struct {
	users         repo.IUserRepo
	sessions      repo.SessionRepositoryInterface
	identities    repo.IOAuthIdentityRepo
	registrations repo.IRegistrationSessionRepo
	registry      *oauth.Registry
	now           func() time.Time
}

func NewOAuthHandler(
	users repo.IUserRepo,
	sessions repo.SessionRepositoryInterface,
	identities repo.IOAuthIdentityRepo,
	registrations repo.IRegistrationSessionRepo,
	registry *oauth.Registry,
) *OAuthHandler {
	return &OAuthHandler{
		users:         users,
		sessions:      sessions,
		identities:    identities,
		registrations: registrations,
		registry:      registry,
		now:           time.Now,
	}
}

// Exchange handles POST /api/oauth/{provider}/exchange.
func (h *OAuthHandler) Exchange(ctx echo.Context, providerName string) error {
	provider, ok := h.registry.Get(providerName)
	if !ok {
		return apiError(ctx, http.StatusNotFound, "not_found", "Unknown OAuth provider")
	}
	if !provider.Enabled() {
		return apiError(ctx, http.StatusNotFound, "not_found", "OAuth provider is disabled")
	}

	var payload oauth.ExchangePayload
	if err := ctx.Bind(&payload); err != nil {
		return badRequest(ctx, "Invalid request body")
	}

	profile, err := provider.Exchange(ctx.Request().Context(), payload)
	if err != nil {
		switch {
		case errors.Is(err, oauth.ErrInvalidPayload),
			errors.Is(err, oauth.ErrSignatureMismatch),
			errors.Is(err, oauth.ErrPayloadExpired):
			return badRequest(ctx, err.Error())
		case errors.Is(err, oauth.ErrProviderDisabled):
			return apiError(ctx, http.StatusNotFound, "not_found", "OAuth provider is disabled")
		default:
			return internalError(ctx, "OAuth exchange failed")
		}
	}

	identity, err := h.identities.GetByProviderUID(ctx.Request().Context(), provider.Name(), profile.ProviderUID)
	switch {
	case err == nil:
		// Existing linked identity → refresh provider tokens, sign the user in.
		applyProfileToIdentity(identity, profile)
		if err := h.identities.Update(ctx.Request().Context(), identity); err != nil {
			return internalError(ctx, "Failed to update OAuth identity")
		}
		return h.signInExisting(ctx, identity.UserID)

	case errors.Is(err, gorm.ErrRecordNotFound):
		// No identity yet → create a transient registration session.
		regSession := &models.RegistrationSession{
			Provider:    provider.Name(),
			ProviderUID: profile.ProviderUID,
			ExpiresAt:   h.now().UTC().Add(registrationSessionTTL),
		}
		applyProfileToRegistration(regSession, profile)
		if err := h.registrations.Create(ctx.Request().Context(), regSession); err != nil {
			return internalError(ctx, "Failed to create registration session")
		}
		return ctx.JSON(http.StatusOK, h.registrationRequiredResponse(provider.Name(), regSession.ID, profile))

	default:
		return internalError(ctx, "Failed to look up OAuth identity")
	}
}

func (h *OAuthHandler) signInExisting(ctx echo.Context, userID uuid.UUID) error {
	user, err := h.users.GetByID(ctx.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apiError(ctx, http.StatusNotFound, "not_found", "Linked user not found")
		}
		return internalError(ctx, "Failed to load user")
	}

	session := &models.Session{
		UserID:    user.ID,
		IP:        ctx.RealIP(),
		UserAgent: ctx.Request().UserAgent(),
	}
	if err := h.sessions.Create(ctx.Request().Context(), session); err != nil {
		return internalError(ctx, "Failed to create session")
	}

	auth := buildAuthResponse(user, session)
	return ctx.JSON(http.StatusOK, oauthExchangeResponse{
		Status: oauthStatusOK,
		Auth:   &auth,
	})
}

func (h *OAuthHandler) registrationRequiredResponse(providerName string, regSessionID uuid.UUID, profile *oauth.Profile) oauthExchangeResponse {
	token := regSessionID
	prov := providerName
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
	return oauthExchangeResponse{
		Status:            oauthStatusRegistrationRequired,
		RegistrationToken: &token,
		Suggested:         &suggested,
		Provider:          &prov,
	}
}

// CompleteSignUp handles POST /api/oauth/complete-signup.
func (h *OAuthHandler) CompleteSignUp(ctx echo.Context) error {
	var req api.OAuthCompleteSignUpRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}

	if req.Username == "" || req.Email == "" || req.RegistrationToken == openapi_types.UUID(uuid.Nil) {
		return badRequest(ctx, "registration_token, username and email are required")
	}

	regSession, err := h.registrations.GetByID(ctx.Request().Context(), uuid.UUID(req.RegistrationToken))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apiError(ctx, http.StatusNotFound, "not_found", "Registration token not found")
		}
		return internalError(ctx, "Failed to load registration session")
	}
	if h.now().UTC().After(regSession.ExpiresAt) {
		_ = h.registrations.Delete(ctx.Request().Context(), regSession.ID)
		return apiError(ctx, http.StatusNotFound, "not_found", "Registration token expired")
	}

	if exists, err := h.users.ExistsByEmail(ctx.Request().Context(), string(req.Email)); err != nil {
		return internalError(ctx, "Failed to check email uniqueness")
	} else if exists {
		return conflict(ctx, "User with this email already exists")
	}
	if exists, err := h.users.ExistsByUsername(ctx.Request().Context(), req.Username); err != nil {
		return internalError(ctx, "Failed to check username uniqueness")
	} else if exists {
		return conflict(ctx, "User with this username already exists")
	}

	user := &models.User{
		Email:     string(req.Email),
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		UserID:    uuid.New(),
	}
	if err := h.users.Create(ctx.Request().Context(), user); err != nil {
		if col := uniqueConstraintColumn(err); col != "" {
			return conflict(ctx, "User with this "+col+" already exists")
		}
		return internalError(ctx, "Failed to create user")
	}

	identity := &models.OAuthIdentity{
		UserID:       user.ID,
		Provider:     regSession.Provider,
		ProviderUID:  regSession.ProviderUID,
		AccessToken:  regSession.AccessToken,
		RefreshToken: regSession.RefreshToken,
		ExpiresAt:    regSession.TokenExpiresAt,
		RawProfile:   regSession.RawProfile,
	}
	if err := h.identities.Create(ctx.Request().Context(), identity); err != nil {
		if col := uniqueConstraintColumn(err); col != "" {
			return conflict(ctx, "OAuth identity already linked ("+col+")")
		}
		return internalError(ctx, "Failed to create OAuth identity")
	}

	// Best-effort cleanup; registration session is single-use.
	_ = h.registrations.Delete(ctx.Request().Context(), regSession.ID)

	session := &models.Session{
		UserID:    user.ID,
		IP:        ctx.RealIP(),
		UserAgent: ctx.Request().UserAgent(),
	}
	if err := h.sessions.Create(ctx.Request().Context(), session); err != nil {
		return internalError(ctx, "Failed to create session")
	}

	return ctx.JSON(http.StatusCreated, buildAuthResponse(user, session))
}

// applyProfileToIdentity refreshes the provider tokens stored on the linked
// identity. Profile snapshots (email/username/...) live on the User row, so we
// don't touch them here — that would silently mutate user-visible data.
func applyProfileToIdentity(identity *models.OAuthIdentity, profile *oauth.Profile) {
	if profile.AccessToken != "" {
		identity.AccessToken = profile.AccessToken
	}
	if profile.RefreshToken != "" {
		identity.RefreshToken = profile.RefreshToken
	}
	if profile.ExpiresAt != nil {
		identity.ExpiresAt = profile.ExpiresAt
	}
	if len(profile.Raw) > 0 {
		identity.RawProfile = profile.Raw
	}
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

func buildAuthResponse(user *models.User, session *models.Session) oauthAuthResponse {
	return oauthAuthResponse{
		SessionToken: session.ID,
		User: oauthUserResponse{
			Id:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			TgUid:     user.TgUID,
			UserId:    user.UserID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}
}
