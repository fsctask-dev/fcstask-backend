package bot

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"

	"fcstask-backend/internal/config"
	"fcstask-backend/internal/oauth"
)

// fakeSender stands in for *tgbotapi.BotAPI, recording the messages the bot
// would send to Telegram.
type fakeSender struct {
	mu   sync.Mutex
	msgs []tgbotapi.MessageConfig
}

func (s *fakeSender) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m, ok := c.(tgbotapi.MessageConfig); ok {
		s.msgs = append(s.msgs, m)
	}
	return tgbotapi.Message{}, nil
}

func (s *fakeSender) last() (tgbotapi.MessageConfig, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.msgs) == 0 {
		return tgbotapi.MessageConfig{}, 0
	}
	return s.msgs[len(s.msgs)-1], len(s.msgs)
}

// exchangeServer is an httptest backend whose /exchange endpoint returns body.
func exchangeServer(body string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/oauth/telegram/exchange", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	})
	return httptest.NewServer(mux)
}

func startUpdate(chatID, userID int64, username string) tgbotapi.Update {
	return tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "/start",
			From: &tgbotapi.User{ID: userID, UserName: username},
			Chat: &tgbotapi.Chat{ID: chatID},
		},
	}
}

// The data the bot signs must verify against the real TelegramProvider, since
// both sides use the same data-check algorithm and bot token.
func TestLoginData_VerifiesWithProvider(t *testing.T) {
	const token = "TEST_BOT_TOKEN"
	b := New(Config{BotToken: token})

	data := b.loginData(&tgbotapi.User{ID: 42, FirstName: "Ann", LastName: "Lee", UserName: "ann"})

	p := oauth.NewTelegramProvider(config.TelegramOAuthConfig{Enabled: true, BotToken: token, MaxAuthAge: 3600})
	prof, err := p.Exchange(context.Background(), oauth.ExchangePayload{TelegramData: data})

	assert.NoError(t, err)
	assert.Equal(t, "42", prof.ProviderUID)
	assert.Equal(t, "ann", prof.Username)
	assert.Equal(t, "Ann", prof.FirstName)
}

// A mismatched bot token must be rejected by the provider.
func TestLoginData_WrongTokenRejected(t *testing.T) {
	b := New(Config{BotToken: "BOT_TOKEN_A"})
	data := b.loginData(&tgbotapi.User{ID: 1, UserName: "x"})

	p := oauth.NewTelegramProvider(config.TelegramOAuthConfig{Enabled: true, BotToken: "BOT_TOKEN_B", MaxAuthAge: 3600})
	_, err := p.Exchange(context.Background(), oauth.ExchangePayload{TelegramData: data})

	assert.ErrorIs(t, err, oauth.ErrSignatureMismatch)
}

func TestExchange_PostsSignedDataToBackend(t *testing.T) {
	var gotPath string
	var gotData map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		var body struct {
			TelegramData map[string]string `json:"telegram_data"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotData = body.TelegramData
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	b := New(Config{BotToken: "T", BackendURL: srv.URL})
	res, err := b.exchange(context.Background(), &tgbotapi.User{ID: 42, UserName: "ann"})

	assert.NoError(t, err)
	assert.Equal(t, "ok", res.Status)
	assert.Equal(t, "/api/oauth/telegram/exchange", gotPath)
	assert.Equal(t, "42", gotData["id"])
	assert.NotEmpty(t, gotData["hash"])
}

func TestExchange_BackendError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	b := New(Config{BotToken: "T", BackendURL: srv.URL})
	_, err := b.exchange(context.Background(), &tgbotapi.User{ID: 1})

	assert.Error(t, err)
}

func TestHandleUpdate_StartSignsInAndReplies(t *testing.T) {
	srv := exchangeServer(`{"status":"ok"}`)
	defer srv.Close()

	b := New(Config{BotToken: "T", BackendURL: srv.URL})
	snd := &fakeSender{}

	b.handleUpdate(context.Background(), snd, startUpdate(9, 42, "ann"))

	m, count := snd.last()
	assert.Equal(t, 1, count)
	assert.Equal(t, int64(9), m.ChatID)
	assert.Contains(t, m.Text, "Signed in as ann")
}

func TestHandleUpdate_RegistrationRequired(t *testing.T) {
	srv := exchangeServer(`{"status":"registration_required","registration_token":"tok-1"}`)
	defer srv.Close()

	b := New(Config{BotToken: "T", BackendURL: srv.URL})
	snd := &fakeSender{}

	b.handleUpdate(context.Background(), snd, startUpdate(9, 42, "ann"))

	m, _ := snd.last()
	assert.Contains(t, m.Text, "registration_token: tok-1")
}

func TestHandleUpdate_BackendErrorReplies(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	b := New(Config{BotToken: "T", BackendURL: srv.URL})
	snd := &fakeSender{}

	b.handleUpdate(context.Background(), snd, startUpdate(9, 42, "ann"))

	m, count := snd.last()
	assert.Equal(t, 1, count)
	assert.Contains(t, m.Text, "sign-in failed")
}

func TestHandleUpdate_NonStartRepliesHelp(t *testing.T) {
	b := New(Config{BotToken: "T"})
	snd := &fakeSender{}

	update := tgbotapi.Update{Message: &tgbotapi.Message{
		Text: "hello",
		From: &tgbotapi.User{ID: 1},
		Chat: &tgbotapi.Chat{ID: 3},
	}}
	b.handleUpdate(context.Background(), snd, update)

	m, count := snd.last()
	assert.Equal(t, 1, count)
	assert.Equal(t, int64(3), m.ChatID)
	assert.Contains(t, m.Text, "Send /start")
}

func TestHandleUpdate_IgnoresMessagesWithoutSender(t *testing.T) {
	b := New(Config{BotToken: "T"})
	snd := &fakeSender{}

	// No panic and no reply for empty/sender-less updates.
	b.handleUpdate(context.Background(), snd, tgbotapi.Update{})
	b.handleUpdate(context.Background(), snd, tgbotapi.Update{Message: &tgbotapi.Message{Text: "/start"}})

	_, count := snd.last()
	assert.Equal(t, 0, count)
}

func TestOutcomeMessage(t *testing.T) {
	b := New(Config{BotToken: "T"})

	ok := b.outcomeMessage(&tgbotapi.User{UserName: "ann"}, &exchangeResult{Status: "ok"})
	assert.Contains(t, ok, "Signed in as ann")

	tok := "abc-123"
	reg := b.outcomeMessage(&tgbotapi.User{UserName: "ann"}, &exchangeResult{Status: "registration_required", RegistrationToken: &tok})
	assert.Contains(t, reg, "registration_token: abc-123")

	bWithFE := New(Config{BotToken: "T", FrontendURL: "https://app.example.com"})
	regLink := bWithFE.outcomeMessage(&tgbotapi.User{UserName: "ann"}, &exchangeResult{Status: "registration_required", RegistrationToken: &tok})
	assert.Contains(t, regLink, "https://app.example.com/signup?registration_token=abc-123")
}
