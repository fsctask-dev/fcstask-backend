package bot

import (
	"context"
	config "fcstask-monitor-bot/internal/config"
	"net/http"

	gotgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func NewBot(ctx context.Context, cfg *config.Config) (*Bot, error) {
	opts := []gotgbot.Option{
		gotgbot.WithMessageTextHandler("/start", gotgbot.MatchTypeExact, Start),
		gotgbot.WithMessageTextHandler("/stop", gotgbot.MatchTypeExact, Stop),
		gotgbot.WithMessageTextHandler("/status", gotgbot.MatchTypeExact, Status),
		gotgbot.WithMessageTextHandler("/help", gotgbot.MatchTypeExact, Help),
		gotgbot.WithMessageTextHandler("✅ Подписаться", gotgbot.MatchTypeExact, Start),
		gotgbot.WithMessageTextHandler("❌ Отписаться", gotgbot.MatchTypeExact, Stop),
		gotgbot.WithMessageTextHandler("❔ Статус", gotgbot.MatchTypeExact, Status),
		gotgbot.WithMessageTextHandler("📋 Помощь", gotgbot.MatchTypeExact, Help),
	}

	bot, err := gotgbot.New(cfg.BotToken, opts...)
	if err != nil {
		return nil, err
	}

	if err := RegisterMyCommands(ctx, bot); err != nil {
		return nil, err
	}

	return &Bot{
		TgBot: bot,
	}, nil
}

func (bot *Bot) StartPolling(ctx context.Context) {
	go func() {
		bot.TgBot.Start(ctx)
	}()
}

func (bot *Bot) StartWebhook(ctx context.Context) {
	go func() {
		bot.TgBot.StartWebhook(ctx)
	}()
}

func (bot *Bot) WebhookHandler() http.Handler {
	return bot.TgBot.WebhookHandler()
}

func RegisterMyCommands(ctx context.Context, bot *gotgbot.Bot) error {
	_, err := bot.SetMyCommands(ctx, &gotgbot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: "start", Description: "✅ Подписаться"},
			{Command: "stop", Description: "❌ Отписаться"},
			{Command: "status", Description: "❔ Статус"},
			{Command: "help", Description: "📋 Помощь"},
		},
	})
	return err
}
