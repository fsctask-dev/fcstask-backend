package config

import "time"

type ObservabilityConfig struct {
	MetricsAddr     string        `yaml:"metrics_addr"`
	DBStatsInterval time.Duration `yaml:"db_stats_interval"`
}
