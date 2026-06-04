package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/api"
	"fcstask-backend/internal/config"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type OAuthHandler struct {
	oauthService *service.OAuthService
	mailerCfg    config.MailerConfig
}

func NewOAuthHandler(oauthService *service.OAuthService) *OAuthHandler {
	return &OAuthHandler{oauthService: oauthService}
}

// WithMailerConfig supplies the resend cooldown used in the pending response of
// an OAuth-initiated signup.
func (h *OAuthHandler) WithMailerConfig(cfg config.MailerConfig) *OAuthHandler {
	h.mailerCfg = cfg
	return h
}

// OAuthExchange exchanges a provider payload for either a session (existing
// identity) or a registration token (first OAuth login) — 200 OK.
func (h *OAuthHandler) OAuthExchange(ctx echo.Context, provider string) error {
	var req api.OAuthExchangeRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}

	result, err := h.oauthService.OauthExchange(ctx.Request().Context(), service.OauthExchangeInput{
		ProviderName: provider,
		Payload:      exchangePayloadFromAPI(req),
		IP:           ctx.RealIP(),
		UserAgent:    ctx.Request().UserAgent(),
	})
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, oauthExchangeToAPI(result))
}

// OAuthAddExchange links an OAuth identity to the authenticated account (200 OK).
func (h *OAuthHandler) OAuthAddExchange(ctx echo.Context, provider string) error {
	user, ok := ctx.Get(UserContextKey).(*models.User)
	if !ok {
		return unauthorized(ctx, "Not authenticated")
	}

	var req api.OAuthExchangeRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}

	if err := h.oauthService.OauthAddExchange(ctx.Request().Context(), user, service.OauthExchangeInput{
		ProviderName: provider,
		Payload:      exchangePayloadFromAPI(req),
		IP:           ctx.RealIP(),
		UserAgent:    ctx.Request().UserAgent(),
	}); err != nil {
		return serviceError(ctx, err)
	}

	return ctx.NoContent(http.StatusOK)
}

// OAuthUnlink removes an OAuth identity from the authenticated account (200 OK).
func (h *OAuthHandler) OAuthUnlink(ctx echo.Context, provider string) error {
	user, ok := ctx.Get(UserContextKey).(*models.User)
	if !ok {
		return unauthorized(ctx, "Not authenticated")
	}

	if err := h.oauthService.OauthUnlink(ctx.Request().Context(), user, provider); err != nil {
		return serviceError(ctx, err)
	}

	return ctx.NoContent(http.StatusOK)
}

// OAuthCompleteSignUp finishes a registration started by an OAuth exchange: it
// creates a pending email registration awaiting verification (201 Created).
func (h *OAuthHandler) OAuthCompleteSignUp(ctx echo.Context) error {
	var req api.OAuthCompleteSignUpRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}

	reg, err := h.oauthService.OauthComplete(ctx.Request().Context(), service.OauthCompleteInput{
		RegistrationToken: req.RegistrationToken,
		Email:             string(req.Email),
		Username:          req.Username,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
	})
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, signUpPendingToAPI(reg, h.mailerCfg.ResendCooldown))
}
