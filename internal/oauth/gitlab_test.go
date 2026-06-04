package oauth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"fcstask-backend/internal/config"
)

func TestGitLabProvider_Exchange_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			body, _ := io.ReadAll(r.Body)
			assert.Contains(t, string(body), "grant_type=authorization_code")
			assert.Contains(t, string(body), "code=AUTHCODE")
			assert.Contains(t, string(body), "code_verifier=VERIFIER")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "AT",
				"refresh_token": "RT",
				"expires_in":    7200,
				"token_type":    "Bearer",
			})
		case "/api/v4/user":
			assert.Equal(t, "Bearer AT", r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":       42,
				"username": "ivan",
				"email":    "ivan@example.com",
				"name":     "Ivan Petrov",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p := NewGitLabProvider(config.GitLabOAuthConfig{
		Enabled:      true,
		BaseURL:      srv.URL,
		ClientID:     "cid",
		ClientSecret: "csec",
	})
	p.client = srv.Client()

	prof, err := p.Exchange(context.Background(), ExchangePayload{
		Code:         "AUTHCODE",
		RedirectURI:  "http://frontend/cb",
		CodeVerifier: "VERIFIER",
	})
	assert.NoError(t, err)
	assert.Equal(t, "42", prof.ProviderUID)
	assert.Equal(t, "ivan", prof.Username)
	assert.Equal(t, "ivan@example.com", prof.Email)
	assert.Equal(t, "Ivan", prof.FirstName)
	assert.Equal(t, "Petrov", prof.LastName)
	assert.Equal(t, "AT", prof.AccessToken)
	assert.Equal(t, "RT", prof.RefreshToken)
	assert.NotNil(t, prof.ExpiresAt)
}

func TestGitLabProvider_Exchange_TokenError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"bad"}`))
	}))
	defer srv.Close()

	p := NewGitLabProvider(config.GitLabOAuthConfig{Enabled: true, BaseURL: srv.URL})
	p.client = srv.Client()

	_, err := p.Exchange(context.Background(), ExchangePayload{Code: "x", RedirectURI: "y"})
	assert.ErrorIs(t, err, ErrTokenExchangeFailed)
}

func TestGitLabProvider_Exchange_Disabled(t *testing.T) {
	p := NewGitLabProvider(config.GitLabOAuthConfig{Enabled: false})
	_, err := p.Exchange(context.Background(), ExchangePayload{Code: "x", RedirectURI: "y"})
	assert.ErrorIs(t, err, ErrProviderDisabled)
}

func TestGitLabProvider_Exchange_InvalidPayload(t *testing.T) {
	p := NewGitLabProvider(config.GitLabOAuthConfig{Enabled: true, BaseURL: "https://example.com"})
	_, err := p.Exchange(context.Background(), ExchangePayload{Code: "", RedirectURI: ""})
	assert.ErrorIs(t, err, ErrInvalidPayload)
}

func TestSplitName(t *testing.T) {
	tests := []struct {
		in, first, last string
	}{
		{"", "", ""},
		{"Ivan", "Ivan", ""},
		{"Ivan Petrov", "Ivan", "Petrov"},
		{"Ivan Petrov Sidorov", "Ivan", "Petrov Sidorov"},
		{"  spaced ", "spaced", ""},
	}
	for _, tc := range tests {
		first, last := splitName(tc.in)
		assert.Equal(t, tc.first, first)
		assert.Equal(t, tc.last, last)
	}
}
