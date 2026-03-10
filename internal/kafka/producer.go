package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"fcstask/internal/config"
)

// Producer publishes messages to Kafka.
type Producer interface {
	PublishMetric(ctx context.Context, metric Metric) error
	Close() error
}

// Writer abstracts the Kafka writer for tests.
type Writer interface {
	WriteMessages(ctx context.Context, msgs ...kafkago.Message) error
	Close() error
}

type kafkaWriter struct {
	writer *kafkago.Writer
}

func (w *kafkaWriter) WriteMessages(ctx context.Context, msgs ...kafkago.Message) error {
	return w.writer.WriteMessages(ctx, msgs...)
}

func (w *kafkaWriter) Close() error {
	return w.writer.Close()
}

type KafkaProducer struct {
	writer       Writer
	metricsTopic string
}

func NewProducer(cfg config.KafkaConfig) (Producer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, errors.New("kafka brokers list is empty")
	}
	if strings.TrimSpace(cfg.TopicMetrics) == "" {
		return nil, errors.New("kafka metrics topic is empty")
	}

	writer := kafkago.NewWriter(kafkago.WriterConfig{
		Brokers:      cfg.Brokers,
		RequiredAcks: cfg.RequiredAcks,
		CompressionCodec: parseCompression(cfg.Compression),
		BatchTimeout: cfg.BatchTimeout,
		BatchSize:    cfg.BatchSize,
		MaxAttempts:  cfg.MaxAttempts,
		Async:        false,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		Dialer: &kafkago.Dialer{
			Timeout: cfg.DialTimeout,
		},
	})
	writer.AllowAutoTopicCreation = cfg.AllowAutoTopicCreation

	return newProducer(&kafkaWriter{writer: writer}, cfg.TopicMetrics), nil
}

// NewProducerWithWriter creates a producer using a custom writer (for tests).
func NewProducerWithWriter(w Writer, metricsTopic string) (Producer, error) {
	if w == nil {
		return nil, errors.New("kafka writer is nil")
	}
	if strings.TrimSpace(metricsTopic) == "" {
		return nil, errors.New("kafka metrics topic is empty")
	}

	return newProducer(w, metricsTopic), nil
}

func newProducer(w Writer, metricsTopic string) *KafkaProducer {
	return &KafkaProducer{
		writer:       w,
		metricsTopic: metricsTopic,
	}
}

func (p *KafkaProducer) publish(ctx context.Context, topic string, key []byte, payload []byte) error {
	if strings.TrimSpace(topic) == "" {
		return errors.New("kafka topic is empty")
	}
	if p.writer == nil {
		return errors.New("kafka writer is nil")
	}

	msg := kafkago.Message{
		Topic: topic,
		Key:   key,
		Value: payload,
		Time:  time.Now().UTC(),
	}

	return p.writer.WriteMessages(ctx, msg)
}

func (p *KafkaProducer) PublishMetric(ctx context.Context, metric Metric) error {
	if p.metricsTopic == "" {
		return errors.New("kafka metrics topic is empty")
	}
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now().UTC()
	}

	payload, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	return p.publish(ctx, p.metricsTopic, []byte(metric.Name), payload)
}

func (p *KafkaProducer) Close() error {
	if p.writer == nil {
		return fmt.Errorf("kafka producer: writer is nil")
	}
	return p.writer.Close()
}

func parseCompression(value string) kafkago.CompressionCodec {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "gzip":
		return kafkago.Gzip.Codec()
	case "snappy":
		return kafkago.Snappy.Codec()
	case "lz4":
		return kafkago.Lz4.Codec()
	case "zstd":
		return kafkago.Zstd.Codec()
	default:
		return nil
	}
}
