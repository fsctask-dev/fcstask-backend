package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Session  SessionConfig  `yaml:"session"`
}

type SessionConfig struct {
	TTL             time.Duration `yaml:"ttl"`
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

	if cfg.Session.TTL == 0 {
		cfg.Session.TTL = 24 * time.Hour
	}

	if cfg.Session.CleanupInterval == 0 {
		cfg.Session.CleanupInterval = 5 * time.Second
	}

	return &cfg, nil
}
