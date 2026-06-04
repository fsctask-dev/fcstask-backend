package handler

import (
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"fcstask-backend/internal/api"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/oauth"
	"fcstask-backend/internal/service"
)

type authResponse struct {
	SessionToken openapi_types.UUID `json:"session_token"`
	User         api.User           `json:"user"`
}

type sessionResponse struct {
	Id        openapi_types.UUID `json:"id"`
	Ip        string             `json:"ip"`
	UserAgent string             `json:"user_agent"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type sessionWithUserResponse struct {
	sessionResponse
	User api.User `json:"user"`
}

type userWithSessionsResponse struct {
	User     api.User          `json:"user"`
	Sessions []sessionResponse `json:"sessions"`
}

type paginatedSessionsResponse struct {
	Items  []sessionWithUserResponse `json:"items"`
	Total  int64                     `json:"total"`
	Limit  int                       `json:"limit"`
	Offset int                       `json:"offset"`
}

type paginatedUsersWithSessionsResponse struct {
	Items  []userWithSessionsResponse `json:"items"`
	Total  int64                      `json:"total"`
	Limit  int                        `json:"limit"`
	Offset int                        `json:"offset"`
}

func userToAPI(user *models.User) api.User {
	return api.User{
		Id:        user.ID,
		Email:     openapi_types.Email(user.Email),
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		TgUid:     user.TgUID,
		UserId:    openapi_types.UUID(user.UserID),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func authResultToAPI(result *service.AuthResult) authResponse {
	return authResponse{
		SessionToken: openapi_types.UUID(result.Session.ID),
		User:         userToAPI(result.User),
	}
}

func signUpPendingToAPI(reg *models.EmailRegistration, resendCooldown time.Duration) api.SignUpPendingResponse {
	return api.SignUpPendingResponse{
		VerificationToken: openapi_types.UUID(reg.ID),
		ResendAfter:       reg.LastSentAt.Add(resendCooldown),
		ExpiresAt:         reg.ExpiresAt,
	}
}

func passwordResetPendingToAPI(pr *models.PasswordReset, resendCooldown time.Duration) api.PasswordResetPendingResponse {
	return api.PasswordResetPendingResponse{
		ResendAfter: pr.LastSentAt.Add(resendCooldown),
		ExpiresAt:   pr.ExpiresAt,
	}
}

type oauthSuggestedProfileResponse struct {
	Email     *string `json:"email,omitempty"`
	Username  *string `json:"username,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

type oauthExchangeResponse struct {
	Status            string                         `json:"status"`
	Auth              *authResponse                  `json:"auth,omitempty"`
	RegistrationToken *openapi_types.UUID            `json:"registration_token,omitempty"`
	Suggested         *oauthSuggestedProfileResponse `json:"suggested,omitempty"`
	Provider          *string                        `json:"provider,omitempty"`
}

func exchangePayloadFromAPI(req api.OAuthExchangeRequest) oauth.ExchangePayload {
	var payload oauth.ExchangePayload
	if req.Code != nil {
		payload.Code = *req.Code
	}
	if req.RedirectUri != nil {
		payload.RedirectURI = *req.RedirectUri
	}
	if req.CodeVerifier != nil {
		payload.CodeVerifier = *req.CodeVerifier
	}
	if req.TelegramData != nil {
		payload.TelegramData = *req.TelegramData
	}
	return payload
}

func oauthExchangeToAPI(result *service.OauthExchangeResult) oauthExchangeResponse {
	if result.RegistrationRequired {
		reg := result.Registration
		token := openapi_types.UUID(reg.Token)
		provider := reg.Provider
		return oauthExchangeResponse{
			Status:            "registration_required",
			RegistrationToken: &token,
			Provider:          &provider,
			Suggested: &oauthSuggestedProfileResponse{
				Email:     reg.SuggestedProfile.Email,
				Username:  reg.SuggestedProfile.Username,
				FirstName: reg.SuggestedProfile.FirstName,
				LastName:  reg.SuggestedProfile.LastName,
			},
		}
	}

	auth := authResultToAPI(result.Auth)
	return oauthExchangeResponse{
		Status: "ok",
		Auth:   &auth,
	}
}
