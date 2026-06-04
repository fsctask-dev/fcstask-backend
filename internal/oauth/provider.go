package oauth

import (
	"context"
	"errors"
	"net/http"
	"time"
)

var (
	ErrProviderDisabled    = errors.New("oauth provider is disabled")
	ErrInvalidPayload      = errors.New("invalid oauth payload")
	ErrTokenExchangeFailed = errors.New("oauth token exchange failed")
	ErrUserInfoFetchFailed = errors.New("oauth user info fetch failed")
	ErrSignatureMismatch   = errors.New("oauth signature mismatch")
	ErrPayloadExpired      = errors.New("oauth payload expired")
)

// Profile is the normalized user profile returned by every provider.
type Profile struct {
	ProviderUID  string
	Email        string
	Username     string
	FirstName    string
	LastName     string
	AccessToken  string
	RefreshToken string
	ExpiresAt    *time.Time
	Raw          []byte
}

// ExchangePayload carries provider-specific request data from the frontend.
// Each provider only reads the fields it needs.
type ExchangePayload struct {
	Code         string            `json:"code,omitempty"`
	RedirectURI  string            `json:"redirect_uri,omitempty"`
	CodeVerifier string            `json:"code_verifier,omitempty"`
	IDToken      string            `json:"id_token,omitempty"`
	TelegramData map[string]string `json:"telegram_data,omitempty"`
}

type Provider interface {
	Name() string
	Enabled() bool
	Exchange(ctx context.Context, payload ExchangePayload) (*Profile, error)
}

// Registry indexes providers by name.
type Registry struct {
	providers map[string]Provider
}

func NewRegistry(providers ...Provider) *Registry {
	r := &Registry{providers: make(map[string]Provider, len(providers))}
	for _, p := range providers {
		if p == nil {
			continue
		}
		r.providers[p.Name()] = p
	}
	return r
}

func (r *Registry) Get(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

// httpClient is the shared client for outbound provider calls. Tests override it.
var httpClient = &http.Client{Timeout: 10 * time.Second}
