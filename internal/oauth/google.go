package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"fcstask-backend/internal/config"
	"fcstask-backend/internal/db/model"
)

const (
	googleTokenURL    = "https://oauth2.googleapis.com/token"
	googleUserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"
)

type GoogleProvider struct {
	cfg    config.GoogleOAuthConfig
	client *http.Client
	// overridable in tests
	tokenURL    string
	userInfoURL string
}

func NewGoogleProvider(cfg config.GoogleOAuthConfig) *GoogleProvider {
	return &GoogleProvider{
		cfg:         cfg,
		client:      httpClient,
		tokenURL:    googleTokenURL,
		userInfoURL: googleUserInfoURL,
	}
}

func (p *GoogleProvider) Name() string  { return model.OAuthProviderGoogle }
func (p *GoogleProvider) Enabled() bool { return p.cfg.Enabled }

type googleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

func (p *GoogleProvider) Exchange(ctx context.Context, payload ExchangePayload) (*Profile, error) {
	if !p.cfg.Enabled {
		return nil, ErrProviderDisabled
	}
	if payload.Code == "" || payload.RedirectURI == "" {
		return nil, ErrInvalidPayload
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", p.cfg.ClientID)
	form.Set("client_secret", p.cfg.ClientSecret)
	form.Set("code", payload.Code)
	form.Set("redirect_uri", payload.RedirectURI)
	if payload.CodeVerifier != "" {
		form.Set("code_verifier", payload.CodeVerifier)
	}

	tokReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL,
		strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}
	tokReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokReq.Header.Set("Accept", "application/json")

	tokResp, err := p.client.Do(tokReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTokenExchangeFailed, err)
	}
	defer tokResp.Body.Close()

	body, _ := io.ReadAll(tokResp.Body)
	if tokResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d body %s", ErrTokenExchangeFailed, tokResp.StatusCode, body)
	}

	var token googleTokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("%w: parse token: %v", ErrTokenExchangeFailed, err)
	}
	if token.Error != "" {
		return nil, fmt.Errorf("%w: %s %s", ErrTokenExchangeFailed, token.Error, token.ErrorDesc)
	}

	userReq, err := http.NewRequestWithContext(ctx, http.MethodGet, p.userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build user request: %w", err)
	}
	userReq.Header.Set("Authorization", "Bearer "+token.AccessToken)
	userReq.Header.Set("Accept", "application/json")

	userResp, err := p.client.Do(userReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUserInfoFetchFailed, err)
	}
	defer userResp.Body.Close()

	rawProfile, _ := io.ReadAll(userResp.Body)
	if userResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrUserInfoFetchFailed, userResp.StatusCode)
	}

	var info googleUserInfo
	if err := json.Unmarshal(rawProfile, &info); err != nil {
		return nil, fmt.Errorf("%w: parse user: %v", ErrUserInfoFetchFailed, err)
	}
	if info.Sub == "" {
		return nil, fmt.Errorf("%w: empty sub", ErrUserInfoFetchFailed)
	}

	prof := &Profile{
		ProviderUID:  info.Sub,
		Email:        info.Email,
		FirstName:    info.GivenName,
		LastName:     info.FamilyName,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Raw:          rawProfile,
	}
	if token.ExpiresIn > 0 {
		t := time.Now().UTC().Add(time.Duration(token.ExpiresIn) * time.Second)
		prof.ExpiresAt = &t
	}
	return prof, nil
}
