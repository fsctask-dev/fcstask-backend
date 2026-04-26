package config

type OAuthConfig struct {
	GitLab   GitLabOAuthConfig   `yaml:"gitlab"`
	Google   GoogleOAuthConfig   `yaml:"google"`
	Telegram TelegramOAuthConfig `yaml:"telegram"`
}

type GitLabOAuthConfig struct {
	Enabled      bool   `yaml:"enabled"`
	BaseURL      string `yaml:"base_url"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

type GoogleOAuthConfig struct {
	Enabled      bool   `yaml:"enabled"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

type TelegramOAuthConfig struct {
	Enabled    bool   `yaml:"enabled"`
	BotToken   string `yaml:"bot_token"`
	MaxAuthAge int64  `yaml:"max_auth_age_seconds"`
}
