package config

import "time"

type KafkaConfig struct {
	Brokers                []string      `yaml:"brokers"`
	TopicMetrics           string        `yaml:"topic_metrics"`
	RequiredAcks           int           `yaml:"required_acks"`
	Compression            string        `yaml:"compression"`
	AllowAutoTopicCreation bool          `yaml:"allow_auto_topic_creation"`
	DialTimeout            time.Duration `yaml:"dial_timeout"`
	ReadTimeout            time.Duration `yaml:"read_timeout"`
	WriteTimeout           time.Duration `yaml:"write_timeout"`
	BatchTimeout           time.Duration `yaml:"batch_timeout"`
	BatchSize              int           `yaml:"batch_size"`
	MaxAttempts            int           `yaml:"max_attempts"`
	MinBytes               int           `yaml:"min_bytes"`
	MaxBytes               int           `yaml:"max_bytes"`
}
