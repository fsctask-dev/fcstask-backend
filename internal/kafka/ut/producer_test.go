package kafka_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"fcstask/internal/kafka"
)

type mockWriter struct {
	messages []kafkago.Message
	err      error
}

func (w *mockWriter) WriteMessages(_ context.Context, msgs ...kafkago.Message) error {
	if w.err != nil {
		return w.err
	}
	w.messages = append(w.messages, msgs...)
	return nil
}

func (w *mockWriter) Close() error {
	return nil
}

func TestProducerPublishMetric(t *testing.T) {
	writer := &mockWriter{}
	producer, err := kafka.NewProducerWithWriter(writer, "metrics")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	metric := kafka.NewMetric("requests_total", 42, map[string]string{
		"route": "/v1/echo",
	}, time.Time{})

	if err := producer.PublishMetric(context.Background(), metric); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(writer.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(writer.messages))
	}

	msg := writer.messages[0]
	if msg.Topic != "metrics" {
		t.Fatalf("expected topic %q, got %q", "metrics", msg.Topic)
	}

	var payload kafka.Metric
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}

	if payload.Name != metric.Name || payload.Value != metric.Value {
		t.Fatalf("unexpected payload: %+v", payload)
	}

	if payload.Timestamp.IsZero() {
		t.Fatalf("expected timestamp to be set")
	}
}

func TestProducerPublishErrors(t *testing.T) {
	writer := &mockWriter{err: errors.New("boom")}
	producer, err := kafka.NewProducerWithWriter(writer, "metrics")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	metric := kafka.NewMetric("test", 1, nil, time.Time{})
	if err := producer.PublishMetric(context.Background(), metric); err == nil {
		t.Fatalf("expected error")
	}
}

func TestProducerPublishMetricUsesTimestamp(t *testing.T) {
	writer := &mockWriter{}
	producer, err := kafka.NewProducerWithWriter(writer, "metrics")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := time.Date(2025, 1, 2, 3, 4, 5, 6, time.UTC)
	metric := kafka.NewMetric("latency_ms", 9.5, nil, expected)

	if err := producer.PublishMetric(context.Background(), metric); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var payload kafka.Metric
	if err := json.Unmarshal(writer.messages[0].Value, &payload); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}

	if !payload.Timestamp.Equal(expected) {
		t.Fatalf("expected timestamp %v, got %v", expected, payload.Timestamp)
	}
}
