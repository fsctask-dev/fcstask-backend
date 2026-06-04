package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"fcstask-backend/internal/config"
	"fcstask-backend/internal/db/model"
)

const defaultTelegramAuthAge int64 = 24 * 60 * 60

type TelegramProvider struct {
	cfg config.TelegramOAuthConfig
	now func() time.Time
}

func NewTelegramProvider(cfg config.TelegramOAuthConfig) *TelegramProvider {
	return &TelegramProvider{cfg: cfg, now: time.Now}
}

func (p *TelegramProvider) Name() string  { return model.OAuthProviderTelegram }
func (p *TelegramProvider) Enabled() bool { return p.cfg.Enabled }

func (p *TelegramProvider) Exchange(ctx context.Context, payload ExchangePayload) (*Profile, error) {
	if !p.cfg.Enabled {
		return nil, ErrProviderDisabled
	}
	if len(payload.TelegramData) == 0 {
		return nil, ErrInvalidPayload
	}
	if p.cfg.BotToken == "" {
		return nil, fmt.Errorf("%w: bot token not configured", ErrProviderDisabled)
	}

	data := payload.TelegramData
	hashHex, ok := data["hash"]
	if !ok || hashHex == "" {
		return nil, fmt.Errorf("%w: missing hash", ErrInvalidPayload)
	}

	id := data["id"]
	if id == "" {
		return nil, fmt.Errorf("%w: missing id", ErrInvalidPayload)
	}

	authDateStr := data["auth_date"]
	if authDateStr == "" {
		return nil, fmt.Errorf("%w: missing auth_date", ErrInvalidPayload)
	}
	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: bad auth_date", ErrInvalidPayload)
	}

	expected := telegramLoginHash(p.cfg.BotToken, data)

	if !hmac.Equal([]byte(strings.ToLower(hashHex)), []byte(expected)) {
		return nil, ErrSignatureMismatch
	}

	maxAge := p.cfg.MaxAuthAge
	if maxAge <= 0 {
		maxAge = defaultTelegramAuthAge
	}
	if p.now().Unix()-authDate > maxAge {
		return nil, ErrPayloadExpired
	}

	raw, _ := json.Marshal(data)

	return &Profile{
		ProviderUID: id,
		Username:    data["username"],
		FirstName:   data["first_name"],
		LastName:    data["last_name"],
		Raw:         raw,
	}, nil
}

// telegramLoginHash computes the Telegram Login Widget data-check hash over all
// fields except "hash", with HMAC-SHA256 keyed by SHA256(botToken).
func telegramLoginHash(botToken string, data map[string]string) string {
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
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(data[k])
	}

	secret := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secret[:])
	mac.Write([]byte(b.String()))
	return hex.EncodeToString(mac.Sum(nil))
}

// SignTelegramLogin fills in the "hash" field of a Telegram Login data set,
// signing it with the bot token exactly as Telegram's Login Widget would. A
// trusted bot — which receives the already-authenticated user from Telegram —
// uses this to produce telegram_data that TelegramProvider.Exchange accepts.
func SignTelegramLogin(botToken string, data map[string]string) {
	data["hash"] = telegramLoginHash(botToken, data)
}
