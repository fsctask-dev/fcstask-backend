package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"fcstask-backend/internal/config"
)

func TestGoogleProvider_Exchange_Success(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "AT",
			"refresh_token": "RT",
			"expires_in":    3600,
			"token_type":    "Bearer",
		})
	}))
	defer tokenSrv.Close()

	userSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer AT", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"sub":         "google-uid-1",
			"email":       "user@gmail.com",
			"given_name":  "Vasya",
			"family_name": "Pupkin",
		})
	}))
	defer userSrv.Close()

	p := NewGoogleProvider(config.GoogleOAuthConfig{
		Enabled:      true,
		ClientID:     "cid",
		ClientSecret: "csec",
	})
	p.client = tokenSrv.Client()
	p.tokenURL = tokenSrv.URL
	p.userInfoURL = userSrv.URL

	prof, err := p.Exchange(context.Background(), ExchangePayload{
		Code:        "AUTHCODE",
		RedirectURI: "http://frontend/cb",
	})
	assert.NoError(t, err)
	assert.Equal(t, "google-uid-1", prof.ProviderUID)
	assert.Equal(t, "user@gmail.com", prof.Email)
	assert.Equal(t, "Vasya", prof.FirstName)
	assert.Equal(t, "Pupkin", prof.LastName)
	assert.Equal(t, "AT", prof.AccessToken)
	assert.Equal(t, "RT", prof.RefreshToken)
}

func TestGoogleProvider_Exchange_Disabled(t *testing.T) {
	p := NewGoogleProvider(config.GoogleOAuthConfig{Enabled: false})
	_, err := p.Exchange(context.Background(), ExchangePayload{Code: "x", RedirectURI: "y"})
	assert.ErrorIs(t, err, ErrProviderDisabled)
}
