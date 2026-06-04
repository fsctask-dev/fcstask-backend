package config

import "time"

type MailerConfig struct {
	Enabled         bool          `yaml:"enabled"`
	From            string        `yaml:"from"`
	SMTP            SMTPConfig    `yaml:"smtp"`
	CodeTTL         time.Duration `yaml:"code_ttl"`
	ResendCooldown  time.Duration `yaml:"resend_cooldown"`
	MaxAttempts     int           `yaml:"max_attempts"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
}

type SMTPConfig struct {
	SSL      bool   `yaml:"ssl"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	PoolSize int    `yaml:"pool_size"`
}
