package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"fcstask-backend/internal/config"
)

func signTelegram(t *testing.T, botToken string, data map[string]string) string {
	t.Helper()
	keys := make([]string, 0, len(data))
	for k := range data {
		if k == "hash" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(k + "=" + data[k])
	}
	secret := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secret[:])
	mac.Write([]byte(b.String()))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestTelegramProvider_Exchange_Valid(t *testing.T) {
	cfg := config.TelegramOAuthConfig{Enabled: true, BotToken: "TEST_BOT_TOKEN", MaxAuthAge: 3600}
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	data := map[string]string{
		"id":         "12345",
		"first_name": "Ivan",
		"last_name":  "Ivanov",
		"username":   "ivan_iv",
		"auth_date":  strconv.FormatInt(now.Unix()-30, 10),
	}
	data["hash"] = signTelegram(t, cfg.BotToken, data)

	p := NewTelegramProvider(cfg)
	p.now = func() time.Time { return now }

	prof, err := p.Exchange(context.Background(), ExchangePayload{TelegramData: data})
	assert.NoError(t, err)
	assert.Equal(t, "12345", prof.ProviderUID)
	assert.Equal(t, "ivan_iv", prof.Username)
	assert.Equal(t, "Ivan", prof.FirstName)
	assert.Equal(t, "Ivanov", prof.LastName)
}

func TestTelegramProvider_Exchange_TamperedHash(t *testing.T) {
	cfg := config.TelegramOAuthConfig{Enabled: true, BotToken: "TEST_BOT_TOKEN", MaxAuthAge: 3600}
	now := time.Now()

	data := map[string]string{
		"id":         "12345",
		"first_name": "Ivan",
		"auth_date":  strconv.FormatInt(now.Unix(), 10),
	}
	data["hash"] = signTelegram(t, cfg.BotToken, data)
	// tamper: caller changes id after signing
	data["id"] = "99999"

	p := NewTelegramProvider(cfg)
	p.now = func() time.Time { return now }

	_, err := p.Exchange(context.Background(), ExchangePayload{TelegramData: data})
	assert.ErrorIs(t, err, ErrSignatureMismatch)
}

func TestTelegramProvider_Exchange_Expired(t *testing.T) {
	cfg := config.TelegramOAuthConfig{Enabled: true, BotToken: "TEST_BOT_TOKEN", MaxAuthAge: 60}
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	data := map[string]string{
		"id":        "12345",
		"auth_date": strconv.FormatInt(now.Unix()-3600, 10),
	}
	data["hash"] = signTelegram(t, cfg.BotToken, data)

	p := NewTelegramProvider(cfg)
	p.now = func() time.Time { return now }

	_, err := p.Exchange(context.Background(), ExchangePayload{TelegramData: data})
	assert.ErrorIs(t, err, ErrPayloadExpired)
}

func TestTelegramProvider_Exchange_Disabled(t *testing.T) {
	cfg := config.TelegramOAuthConfig{Enabled: false, BotToken: "TEST_BOT_TOKEN"}
	p := NewTelegramProvider(cfg)
	_, err := p.Exchange(context.Background(), ExchangePayload{TelegramData: map[string]string{"id": "1"}})
	assert.ErrorIs(t, err, ErrProviderDisabled)
}

func TestTelegramProvider_Exchange_MissingFields(t *testing.T) {
	cfg := config.TelegramOAuthConfig{Enabled: true, BotToken: "TEST_BOT_TOKEN", MaxAuthAge: 60}
	p := NewTelegramProvider(cfg)

	cases := []struct {
		name string
		data map[string]string
	}{
		{"empty", map[string]string{}},
		{"missing hash", map[string]string{"id": "1", "auth_date": "1"}},
		{"missing id", map[string]string{"hash": "x", "auth_date": "1"}},
		{"missing auth_date", map[string]string{"hash": "x", "id": "1"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := p.Exchange(context.Background(), ExchangePayload{TelegramData: tc.data})
			if !errors.Is(err, ErrInvalidPayload) {
				t.Fatalf("want ErrInvalidPayload, got %v", err)
			}
		})
	}
}
