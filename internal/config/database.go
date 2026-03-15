package config

import "time"

type DatabaseConfig struct {
	Host            string          `yaml:"host"`
	Port            int             `yaml:"port"`
	Username        string          `yaml:"username"`
	Password        string          `yaml:"password"`
	Database        string          `yaml:"database"`
	SSLMode         string          `yaml:"ssl_mode"`
	MaxOpenConns    int             `yaml:"max_open_conns"`
	MaxIdleConns    int             `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration   `yaml:"conn_max_lifetime"`
	Replicas        []ReplicaConfig `yaml:"replicas,omitempty"`
}

type ReplicaConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"ssl_mode"`
}
