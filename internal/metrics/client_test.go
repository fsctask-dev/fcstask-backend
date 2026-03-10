package metrics_test

import (
	"context"
	"testing"

	"fcstask/internal/kafka"
	"fcstask/internal/metrics"
)

type mockProducer struct {
	metrics []kafka.Metric
	err     error
}

func (m *mockProducer) PublishMetric(ctx context.Context, metric kafka.Metric) error {
	if m.err != nil {
		return m.err
	}
	m.metrics = append(m.metrics, metric)
	return nil
}

func (m *mockProducer) Close() error {
	return nil
}

func TestKafkaMetricsClient_Inc(t *testing.T) {
	mock := &mockProducer{}
	ctx := context.Background()

	client := metrics.NewKafkaMetricsClient(mock, "test", ctx)

	client.Inc("test_counter", map[string]string{"env": "test"})

	if len(mock.metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(mock.metrics))
	}

	metric := mock.metrics[0]
	if metric.Name != "test_counter" {
		t.Errorf("expected name 'test_counter', got %q", metric.Name)
	}
	if metric.Value != 1 {
		t.Errorf("expected value 1, got %f", metric.Value)
	}
	if metric.Tags["env"] != "test" {
		t.Errorf("expected tag env=test, got %v", metric.Tags)
	}
	if metric.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
}
