package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Kafka  KafkaConfig  `yaml:"kafka"`
}

type ServerConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type KafkaConfig struct {
	Brokers      []string      `yaml:"brokers"`
	TopicMetrics string        `yaml:"topic_metrics"`
	RequiredAcks int           `yaml:"required_acks"`
	Compression  string        `yaml:"compression"`
	AllowAutoTopicCreation bool `yaml:"allow_auto_topic_creation"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	BatchTimeout time.Duration `yaml:"batch_timeout"`
	BatchSize    int           `yaml:"batch_size"`
	MaxAttempts  int           `yaml:"max_attempts"`
	MinBytes     int           `yaml:"min_bytes"`
	MaxBytes     int           `yaml:"max_bytes"`
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
