//go:build integration

package kafka_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	tcKafka "github.com/testcontainers/testcontainers-go/modules/kafka"

	"fcstask/internal/config"
	"fcstask/internal/kafka"
)

func TestProducerPublishMetricIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	kafkaContainer, err := tcKafka.RunContainer(ctx)
	if err != nil {
		t.Fatalf("failed to start kafka container: %v", err)
	}
	t.Cleanup(func() {
		_ = kafkaContainer.Terminate(context.Background())
	})

	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		t.Fatalf("failed to get brokers: %v", err)
	}

	if err := ensureTopic(ctx, brokers[0], "metrics"); err != nil {
		t.Fatalf("failed to ensure topic: %v", err)
	}

	cfg := config.KafkaConfig{
		Brokers:                brokers,
		TopicMetrics:           "metrics",
		RequiredAcks:           -1,
		Compression:            "snappy",
		AllowAutoTopicCreation: true,
		DialTimeout:            5 * time.Second,
		ReadTimeout:            5 * time.Second,
		WriteTimeout:           5 * time.Second,
		BatchTimeout:           10 * time.Millisecond,
		BatchSize:              1,
		MaxAttempts:            3,
	}

	producer, err := kafka.NewProducer(cfg)
	if err != nil {
		t.Fatalf("failed to create producer: %v", err)
	}
	t.Cleanup(func() {
		_ = producer.Close()
	})

	metric := kafka.NewMetric("requests_total", 7, map[string]string{
		"route": "/v1/echo",
	}, time.Time{})

	if err := producer.PublishMetric(ctx, metric); err != nil {
		t.Fatalf("failed to publish metric: %v", err)
	}

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  brokers,
		Topic:    "metrics",
		GroupID:  "metrics-test",
		MinBytes: 1,
		MaxBytes: 10 * 1024 * 1024,
	})
	defer reader.Close()

	msg, err := reader.ReadMessage(ctx)
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var payload kafka.Metric
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		t.Fatalf("failed to unmarshal metric: %v", err)
	}

	if payload.Name != metric.Name || payload.Value != metric.Value {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func ensureTopic(ctx context.Context, broker string, topic string) error {
	conn, err := kafkago.DialContext(ctx, "tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	err = conn.CreateTopics(kafkago.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	if err == nil {
		return nil
	}

	if strings.Contains(strings.ToLower(err.Error()), "already exists") {
		return nil
	}

	return err
}
