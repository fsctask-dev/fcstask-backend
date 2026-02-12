package recovery

import (
	"os"

	"gopkg.in/yaml.v3"
)

type TargetConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"ssl_mode"`
}

type RestoreConfig struct {
	Target            TargetConfig `yaml:"target"`
	BackupRoot        string       `yaml:"backup_root"`
	DropDatabase      bool         `yaml:"drop_database"`
	Jobs              int          `yaml:"jobs"`
	WalDestinationDir string       `yaml:"wal_destination_dir"`
	LogFile           string       `yaml:"log_file"`
}

func DefaultRestoreConfig() RestoreConfig {
	return RestoreConfig{
		Target: TargetConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "",
			Database: "postgres",
			SSLMode:  "disable",
		},
		BackupRoot:        "/var/backups/postgres",
		DropDatabase:      true,
		Jobs:              4,
		WalDestinationDir: "",
		LogFile:           "",
	}
}

func LoadRestoreConfig(path string) (*RestoreConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg RestoreConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
