package config

import (
	"fmt"
	"os"
	"time"

	"fcstask-backend/backup-service/internal/sqlutil"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Source        ConnConfig    `yaml:"source"`
	RestoreTarget ConnConfig    `yaml:"restore_target"`
	Backup        BackupConfig  `yaml:"backup"`
	PITR          PITRConfig    `yaml:"pitr"`
	WAL           WALConfig     `yaml:"wal"`
	Cron          CronConfig    `yaml:"cron"`
	Restore       RestoreConfig `yaml:"restore"`
	Logging       LoggingConfig `yaml:"logging"`
	Health        HealthConfig  `yaml:"health"`
}

type HealthConfig struct {
	Addr string `yaml:"addr"`
}

type ConnConfig struct {
	Host                  string `yaml:"host"`
	Port                  int    `yaml:"port"`
	User                  string `yaml:"user"`
	Password              string `yaml:"password"`
	Database              string `yaml:"database"`
	SSLMode               string `yaml:"ssl_mode"`
	ConnectTimeoutSeconds int    `yaml:"connect_timeout_seconds"`
}

type BackupConfig struct {
	OutputDir             string `yaml:"output_dir"`
	MinFreeSpaceGB        int    `yaml:"min_free_space_gb"`
	SplitSizeMB           int    `yaml:"split_size_mb"`
	RetentionDays         int    `yaml:"retention_days"`
	PgDumpJobs            int    `yaml:"pg_dump_jobs"`
	FullBackupEvery       int    `yaml:"full_backup_every"`
	CommandTimeoutMinutes int    `yaml:"command_timeout_minutes"`
}

type PITRConfig struct {
	Enabled          bool     `yaml:"enabled"`
	Schemas          []string `yaml:"schemas"`
	CreatedAtColumns []string `yaml:"created_at_columns"`
	UpdatedAtColumns []string `yaml:"updated_at_columns"`
	DeletedAtColumns []string `yaml:"deleted_at_columns"`
}

type WALConfig struct {
	Enabled        bool   `yaml:"enabled"`
	BaseBackupDir  string `yaml:"base_backup_dir"`
	ArchiveDir     string `yaml:"archive_dir"`
	SlotName       string `yaml:"slot_name"`
	FastCheckpoint bool   `yaml:"fast_checkpoint"`
	RetentionDays  int    `yaml:"retention_days"`

	Replication ConnConfig `yaml:"replication"`

	RestoreDataDir string `yaml:"restore_data_dir"`
	PgCtlPath      string `yaml:"pg_ctl_path"`
	RestorePort    int    `yaml:"restore_port"`
}

type CronConfig struct {
	Schedule          string `yaml:"schedule"`
	RunInitialOnStart bool   `yaml:"run_initial_on_start"`
}

type RestoreConfig struct {
	Jobs         int  `yaml:"jobs"`
	DropDatabase bool `yaml:"drop_database"`
}

type LoggingConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAgeDays int    `yaml:"max_age_days"`
	Stdout     bool   `yaml:"stdout"`
}

func (b BackupConfig) CommandTimeout() time.Duration {
	if b.CommandTimeoutMinutes <= 0 {
		return 2 * time.Hour
	}
	return time.Duration(b.CommandTimeoutMinutes) * time.Minute
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	cfg := defaults()

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &cfg, nil
}

func defaults() Config {
	return Config{
		Backup: BackupConfig{
			MinFreeSpaceGB:        10,
			SplitSizeMB:           100,
			RetentionDays:         30,
			PgDumpJobs:            4,
			FullBackupEvery:       7,
			CommandTimeoutMinutes: 120,
		},
		PITR: PITRConfig{
			Enabled:          true,
			Schemas:          []string{"public"},
			CreatedAtColumns: []string{"created_at"},
			UpdatedAtColumns: []string{"updated_at"},
			DeletedAtColumns: []string{"deleted_at"},
		},
		WAL: WALConfig{
			Enabled:        false,
			SlotName:       "backup_service_slot",
			FastCheckpoint: true,
			RetentionDays:  30,
			RestorePort:    5433,
		},
		Cron: CronConfig{Schedule: "0 2 * * *"},
		Restore: RestoreConfig{
			Jobs:         4,
			DropDatabase: true,
		},
		Logging: LoggingConfig{
			Level:      "info",
			MaxSizeMB:  100,
			MaxBackups: 10,
			MaxAgeDays: 30,
			Stdout:     true,
		},
	}
}

func (c *Config) validate() error {
	if err := validateConn("source", c.Source); err != nil {
		return err
	}
	if c.Backup.OutputDir == "" {
		return fmt.Errorf("backup.output_dir is required")
	}
	if c.Backup.PgDumpJobs <= 0 {
		c.Backup.PgDumpJobs = 1
	}
	if c.Backup.FullBackupEvery <= 0 {
		c.Backup.FullBackupEvery = 1
	}
	if c.Restore.Jobs <= 0 {
		c.Restore.Jobs = 1
	}

	if c.RestoreTarget.Host != "" {
		if err := validateConn("restore_target", c.RestoreTarget); err != nil {
			return err
		}
	}
	if c.PITR.Enabled && len(c.PITR.CreatedAtColumns) == 0 && len(c.PITR.UpdatedAtColumns) == 0 {
		return fmt.Errorf("pitr.enabled is true but no created_at/updated_at columns configured")
	}
	if c.PITR.Enabled {
		for _, s := range c.PITR.Schemas {
			if err := sqlutil.ValidateIdentifier("pitr schema", s); err != nil {
				return err
			}
		}
		cols := append(append(append([]string{}, c.PITR.CreatedAtColumns...), c.PITR.UpdatedAtColumns...), c.PITR.DeletedAtColumns...)
		for _, col := range cols {
			if err := sqlutil.ValidateIdentifier("pitr column", col); err != nil {
				return err
			}
		}
	}
	if c.WAL.Enabled {
		if c.WAL.BaseBackupDir == "" {
			return fmt.Errorf("wal.base_backup_dir is required when wal.enabled is true")
		}
		if c.WAL.ArchiveDir == "" {
			return fmt.Errorf("wal.archive_dir is required when wal.enabled is true")
		}
		if c.WAL.Replication.Host == "" {
			return fmt.Errorf("wal.replication.host is required when wal.enabled is true")
		}
		if c.WAL.Replication.Port <= 0 {
			return fmt.Errorf("wal.replication.port is invalid: %d", c.WAL.Replication.Port)
		}
		if c.WAL.Replication.User == "" {
			return fmt.Errorf("wal.replication.user is required when wal.enabled is true")
		}
		if c.WAL.SlotName == "" {
			c.WAL.SlotName = "backup_service_slot"
		}
	}
	return nil
}

func validateConn(name string, c ConnConfig) error {
	if c.Host == "" {
		return fmt.Errorf("%s.host is required", name)
	}
	if c.Port <= 0 {
		return fmt.Errorf("%s.port is invalid: %d", name, c.Port)
	}
	if c.User == "" {
		return fmt.Errorf("%s.user is required", name)
	}
	if c.Database == "" {
		return fmt.Errorf("%s.database is required", name)
	}
	return nil
}
