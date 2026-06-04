package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fcstask-backend/telegram-bot/pkg/bot"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("telegram-bot: TELEGRAM_BOT_TOKEN is required (must match the backend's oauth.telegram.bot_token)")
	}

	b := bot.New(bot.Config{
		BotToken:    token,
		BackendURL:  getenv("BACKEND_BASE_URL", "http://localhost:8080"),
		FrontendURL: os.Getenv("FRONTEND_BASE_URL"),
		PollTimeout: 30 * time.Second,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := b.Run(ctx); err != nil {
		log.Fatalf("telegram-bot: %v", err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
