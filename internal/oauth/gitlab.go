package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fcstask-backend/internal/config"
	"fcstask-backend/internal/db/model"
)

type GitLabProvider struct {
	cfg    config.GitLabOAuthConfig
	client *http.Client
}

func NewGitLabProvider(cfg config.GitLabOAuthConfig) *GitLabProvider {
	return &GitLabProvider{cfg: cfg, client: httpClient}
}

func (p *GitLabProvider) Name() string  { return model.OAuthProviderGitLab }
func (p *GitLabProvider) Enabled() bool { return p.cfg.Enabled }

type gitlabTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	CreatedAt    int64  `json:"created_at"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

type gitlabUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

func (p *GitLabProvider) Exchange(ctx context.Context, payload ExchangePayload) (*Profile, error) {
	if !p.cfg.Enabled {
		return nil, ErrProviderDisabled
	}
	if payload.Code == "" || payload.RedirectURI == "" {
		return nil, ErrInvalidPayload
	}

	base := strings.TrimRight(p.cfg.BaseURL, "/")

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", p.cfg.ClientID)
	form.Set("client_secret", p.cfg.ClientSecret)
	form.Set("code", payload.Code)
	form.Set("redirect_uri", payload.RedirectURI)
	if payload.CodeVerifier != "" {
		form.Set("code_verifier", payload.CodeVerifier)
	}

	tokReq, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/oauth/token",
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

	var token gitlabTokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("%w: parse token: %v", ErrTokenExchangeFailed, err)
	}
	if token.Error != "" {
		return nil, fmt.Errorf("%w: %s %s", ErrTokenExchangeFailed, token.Error, token.ErrorDesc)
	}

	userReq, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/v4/user", nil)
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

	var u gitlabUser
	if err := json.Unmarshal(rawProfile, &u); err != nil {
		return nil, fmt.Errorf("%w: parse user: %v", ErrUserInfoFetchFailed, err)
	}
	if u.ID == 0 {
		return nil, fmt.Errorf("%w: empty user id", ErrUserInfoFetchFailed)
	}

	first, last := splitName(u.Name)

	prof := &Profile{
		ProviderUID:  strconv.FormatInt(u.ID, 10),
		Email:        u.Email,
		Username:     u.Username,
		FirstName:    first,
		LastName:     last,
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

func splitName(full string) (first, last string) {
	full = strings.TrimSpace(full)
	if full == "" {
		return "", ""
	}
	parts := strings.SplitN(full, " ", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}
