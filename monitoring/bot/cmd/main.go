package main

import (
	"context"
	tgbot "fcstask-monitor-bot/internal/bot"
	"fcstask-monitor-bot/internal/config"
	database "fcstask-monitor-bot/internal/db"
	"fcstask-monitor-bot/internal/logger"
	server "fcstask-monitor-bot/internal/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to load config")
	}

	if err := logger.InitLogger(cfg.LogLevel, cfg.LogFile, cfg.LogConsole); err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to initialize logger")
	}

	logger.Log.Info().Str("level", cfg.LogLevel).Msg("logger initialized")

	database.InitDB()

	bot, err := tgbot.NewBot(ctx, cfg)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to create bot")
	}

	switch cfg.BotWebhook {
	case true:
		bot.StartWebhook(ctx)
	case false:
		bot.StartPolling(ctx)
	}

	serverFiber := server.NewServer(ctx, bot)
	if err := serverFiber.Run(ctx, cfg); err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to run server")
	}
}
