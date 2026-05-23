package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

func NewConfig() (*Config, error) {
	godotenv.Load()

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		return nil, errors.New("BOT_TOKEN environment variable is not set")
	}

	botWebhook := os.Getenv("BOT_WEBHOOK") == "true"

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		return nil, errors.New("SERVER_PORT environment variable is not set")
	}

	publicURL := os.Getenv("PUBLIC_URL")
	if botWebhook && publicURL == "" {
		return nil, errors.New("PUBLIC_URL environment variable is not set")
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	logFile := os.Getenv("LOG_FILE")
	logConsole := os.Getenv("LOG_CONSOLE") == "true"

	return &Config{
		BotToken:   botToken,
		BotWebhook: botWebhook,
		ServerPort: serverPort,
		PublicURL:  publicURL,
		LogLevel:   logLevel,
		LogFile:    logFile,
		LogConsole: logConsole,
	}, nil
}
