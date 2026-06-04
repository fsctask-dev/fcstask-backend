package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server            ServerConfig            `yaml:"server"`
	Database          DatabaseConfig          `yaml:"database"`
	Session           SessionConfig           `yaml:"session"`
	PasswordReset     PasswordResetConfig     `yaml:"password_reset"`
	EmailRegistration EmailRegistrationConfig `yaml:"email_registration"`
	OAuth             OAuthConfig             `yaml:"oauth"`
	Mailer            MailerConfig            `yaml:"mailer"`
	Observability     ObservabilityConfig     `yaml:"observability"`
}

type SessionConfig struct {
	TTL             time.Duration `yaml:"ttl"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
}

type EmailRegistrationConfig struct {
	TTL             time.Duration `yaml:"ttl"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
}

type PasswordResetConfig struct {
	TTL             time.Duration `yaml:"ttl"`
	MaxAttempts     int           `yaml:"max_attempts"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
}

type ServerConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	for i := range cfg.Database.Replicas {
		if cfg.Database.Replicas[i].SSLMode == "" {
			cfg.Database.Replicas[i].SSLMode = "disable"
		}
	}
	if cfg.Database.MaxOpenConns <= 0 {
		cfg.Database.MaxOpenConns = 25
	}
	if cfg.Database.MaxIdleConns <= 0 {
		cfg.Database.MaxIdleConns = 25
	}
	if cfg.Database.ConnMaxLifetime <= 0 {
		cfg.Database.ConnMaxLifetime = 5 * time.Minute
	}

	if cfg.Session.TTL == 0 {
		cfg.Session.TTL = 24 * time.Hour
	}

	if cfg.Session.CleanupInterval == 0 {
		cfg.Session.CleanupInterval = 5 * time.Second
	}

	if cfg.Mailer.CodeTTL == 0 {
		cfg.Mailer.CodeTTL = 15 * time.Minute
	}
	if cfg.Mailer.ResendCooldown == 0 {
		cfg.Mailer.ResendCooldown = 60 * time.Second
	}
	if cfg.Mailer.MaxAttempts <= 0 {
		cfg.Mailer.MaxAttempts = 5
	}
	if cfg.Mailer.CleanupInterval == 0 {
		cfg.Mailer.CleanupInterval = 5 * time.Minute
	}

	if cfg.EmailRegistration.TTL == 0 {
		cfg.EmailRegistration.TTL = cfg.Mailer.CodeTTL
	}
	if cfg.EmailRegistration.CleanupInterval == 0 {
		cfg.EmailRegistration.CleanupInterval = cfg.Mailer.CleanupInterval
	}
	if cfg.PasswordReset.TTL == 0 {
		cfg.PasswordReset.TTL = cfg.Mailer.CodeTTL
	}
	if cfg.PasswordReset.CleanupInterval == 0 {
		cfg.PasswordReset.CleanupInterval = cfg.Mailer.CleanupInterval
	}
	if cfg.OAuth.RegistractionTTL == 0 {
		cfg.OAuth.RegistractionTTL = 15 * time.Minute
	}
	if cfg.OAuth.CleanupInterval == 0 {
		cfg.OAuth.CleanupInterval = cfg.Mailer.CleanupInterval
	}

	if cfg.Observability.MetricsAddr == "" {
		cfg.Observability.MetricsAddr = ":8081"
	}
	if cfg.Observability.DBStatsInterval <= 0 {
		cfg.Observability.DBStatsInterval = 15 * time.Second
	}

	return &cfg, nil
}
