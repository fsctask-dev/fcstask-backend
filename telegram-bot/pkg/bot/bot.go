// Package bot implements a Telegram bot that authenticates users via the
// backend's Telegram OAuth provider.
//
// Flow: a user opens the bot and sends /start. Telegram delivers the message
// together with the (already authenticated) sender profile. The bot rebuilds
// the Telegram Login data set, signs it with the bot token exactly like the
// Login Widget would, and POSTs it to the backend's
// POST /api/oauth/{provider}/exchange endpoint. The backend verifies the
// signature and either signs the user in or asks them to finish registration.
// The bot reports the outcome back in the chat.
//
// The Telegram transport (long polling, sending messages) is handled by
// github.com/go-telegram-bot-api/telegram-bot-api/v5.
package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/oauth"
)

// Config holds the bot's runtime configuration.
type Config struct {
	// BotToken is the Telegram bot token (from @BotFather). Also used to sign
	// the Telegram Login data, so it must match the backend's
	// oauth.telegram.bot_token.
	BotToken string
	// BackendURL is the base URL of the fcstask backend, e.g.
	// http://host.docker.internal:8080.
	BackendURL string
	// FrontendURL, if set, is used to build a complete-signup link for new users.
	FrontendURL string
	// PollTimeout is the long-poll timeout for getUpdates.
	PollTimeout time.Duration
	// HTTPTimeout bounds backend exchange requests.
	HTTPTimeout time.Duration
}

// sender is the subset of *tgbotapi.BotAPI the bot uses to reply. It lets tests
// stub out the Telegram transport.
type sender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

// Bot long-polls Telegram and bridges Telegram logins to the backend.
type Bot struct {
	cfg    Config
	client *http.Client
}

func New(cfg Config) *Bot {
	if cfg.PollTimeout <= 0 {
		cfg.PollTimeout = 30 * time.Second
	}
	if cfg.HTTPTimeout <= 0 {
		cfg.HTTPTimeout = 30 * time.Second
	}
	return &Bot{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.HTTPTimeout},
	}
}

// Run connects to Telegram and long-polls until ctx is cancelled.
func (b *Bot) Run(ctx context.Context) error {
	if b.cfg.BotToken == "" {
		return fmt.Errorf("bot: token is required")
	}

	api, err := tgbotapi.NewBotAPI(b.cfg.BotToken)
	if err != nil {
		return fmt.Errorf("bot: connect to telegram: %w", err)
	}
	log.Printf("bot: started as @%s, backend=%s", api.Self.UserName, b.cfg.BackendURL)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = int(b.cfg.PollTimeout / time.Second)
	u.AllowedUpdates = []string{"message"}
	updates := api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			api.StopReceivingUpdates()
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			b.handleUpdate(ctx, api, update)
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, snd sender, update tgbotapi.Update) {
	msg := update.Message
	if msg == nil || msg.From == nil {
		return
	}

	if !strings.HasPrefix(msg.Text, "/start") {
		b.reply(snd, msg.Chat.ID, "Send /start to sign in with Telegram.")
		return
	}

	result, err := b.exchange(ctx, msg.From)
	if err != nil {
		log.Printf("bot: exchange for tg id %d failed: %v", msg.From.ID, err)
		b.reply(snd, msg.Chat.ID, "Sorry, sign-in failed. Please try again later.")
		return
	}

	b.reply(snd, msg.Chat.ID, b.outcomeMessage(msg.From, result))
}

// loginData builds and signs the Telegram Login data set for the sender. The
// bot vouches for the user with auth_date = now, since Telegram already
// authenticated the incoming message.
func (b *Bot) loginData(from *tgbotapi.User) map[string]string {
	data := map[string]string{
		"id":        strconv.FormatInt(from.ID, 10),
		"auth_date": strconv.FormatInt(time.Now().Unix(), 10),
	}
	if from.FirstName != "" {
		data["first_name"] = from.FirstName
	}
	if from.LastName != "" {
		data["last_name"] = from.LastName
	}
	if from.UserName != "" {
		data["username"] = from.UserName
	}
	oauth.SignTelegramLogin(b.cfg.BotToken, data)
	return data
}

type exchangeResult struct {
	Status            string  `json:"status"`
	RegistrationToken *string `json:"registration_token,omitempty"`
}

func (b *Bot) exchange(ctx context.Context, from *tgbotapi.User) (*exchangeResult, error) {
	payload := map[string]any{"telegram_data": b.loginData(from)}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoint := strings.TrimRight(b.cfg.BackendURL, "/") + "/api/oauth/" + model.OAuthProviderTelegram + "/exchange"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("backend status %d: %s", resp.StatusCode, respBody)
	}

	var out exchangeResult
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (b *Bot) outcomeMessage(from *tgbotapi.User, result *exchangeResult) string {
	name := from.UserName
	if name == "" {
		name = from.FirstName
	}

	switch result.Status {
	case "ok":
		return fmt.Sprintf("✅ Signed in as %s.", name)
	case "registration_required":
		if b.cfg.FrontendURL != "" && result.RegistrationToken != nil {
			link := strings.TrimRight(b.cfg.FrontendURL, "/") + "/signup?registration_token=" + url.QueryEscape(*result.RegistrationToken)
			return "👋 Looks like you're new here. Finish signup: " + link
		}
		token := ""
		if result.RegistrationToken != nil {
			token = "\nregistration_token: " + *result.RegistrationToken
		}
		return "👋 Looks like you're new here. Finish your registration in the app." + token
	default:
		return "Done."
	}
}

func (b *Bot) reply(snd sender, chatID int64, text string) {
	if _, err := snd.Send(tgbotapi.NewMessage(chatID, text)); err != nil {
		log.Printf("bot: send error: %v", err)
	}
}
