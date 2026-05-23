package config

type Config struct {
	BotToken   string `env:"BOT_TOKEN,required"`
	BotWebhook bool   `env:"BOT_WEBHOOK,required"`
	ServerPort string `env:"SERVER_PORT,required"`
	PublicURL  string `env:"PUBLIC_URL,required"`
	LogLevel   string `env:"LOG_LEVEL"`
	LogFile    string `env:"LOG_FILE"`
	LogConsole bool   `env:"LOG_CONSOLE"`
}
