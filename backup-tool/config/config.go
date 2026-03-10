package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Postgres PostgresConfig `yaml:"postgres"`
	Backup   BackupConfig   `yaml:"backup"`
	Cron     CronConfig     `yaml:"cron"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"ssl_mode"`
}

type BackupConfig struct {
	OutputDir      string `yaml:"output_dir"`
	MinFreeSpaceGB int    `yaml:"min_free_space_gb"`
	SplitSizeMB    int    `yaml:"split_size_mb"`
	RetentionDays  int    `yaml:"retention_days"`
}

type CronConfig struct {
	Schedule string `yaml:"schedule"`
}

type LoggingConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAgeDays int    `yaml:"max_age_days"`
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

	return &cfg, nil
}
