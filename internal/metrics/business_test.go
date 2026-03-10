package metrics_test

import (
	"context"
	"testing"

	"fcstask/internal/kafka"
	"fcstask/internal/metrics"
)

// mockProducer для бизнес-метрик
type businessMockProducer struct {
	metrics []kafka.Metric
}

func (m *businessMockProducer) PublishMetric(ctx context.Context, metric kafka.Metric) error {
	m.metrics = append(m.metrics, metric)
	return nil
}

func (m *businessMockProducer) Close() error {
	return nil
}

func TestBusinessMetrics_TaskSubmitted(t *testing.T) {
	mock := &businessMockProducer{}
	ctx := context.Background()
	client := metrics.NewKafkaMetricsClient(mock, "test", ctx)
	business := metrics.NewBusinessMetrics(client)

	business.TaskSubmitted("algorithms-101", "hw1")

	if len(mock.metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(mock.metrics))
	}

	metric := mock.metrics[0]
	if metric.Name != "tasks_submitted_total" {
		t.Errorf("expected name %q, got %q", "tasks_submitted_total", metric.Name)
	}
	if metric.Value != 1 {
		t.Errorf("expected value 1, got %f", metric.Value)
	}
	if metric.Tags["course"] != "algorithms-101" {
		t.Errorf("expected course algorithms-101, got %v", metric.Tags["course"])
	}
	if metric.Tags["assignment"] != "hw1" {
		t.Errorf("expected assignment hw1, got %v", metric.Tags["assignment"])
	}
}

func TestBusinessMetrics_TaskCompleted(t *testing.T) {
	mock := &businessMockProducer{}
	ctx := context.Background()
	client := metrics.NewKafkaMetricsClient(mock, "test", ctx)
	business := metrics.NewBusinessMetrics(client)

	business.TaskCompleted("algorithms-101", "hw1")

	if len(mock.metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(mock.metrics))
	}

	metric := mock.metrics[0]
	if metric.Name != "tasks_completed_total" {
		t.Errorf("expected name %q, got %q", "tasks_completed_total", metric.Name)
	}
	if metric.Tags["status"] != "success" {
		t.Errorf("expected status success, got %v", metric.Tags["status"])
	}
}

func TestBusinessMetrics_TaskFailed(t *testing.T) {
	mock := &businessMockProducer{}
	ctx := context.Background()
	client := metrics.NewKafkaMetricsClient(mock, "test", ctx)
	business := metrics.NewBusinessMetrics(client)

	business.TaskFailed("algorithms-101", "hw1")

	if len(mock.metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(mock.metrics))
	}

	metric := mock.metrics[0]
	if metric.Name != "tasks_failed_total" {
		t.Errorf("expected name %q, got %q", "tasks_failed_total", metric.Name)
	}
	if metric.Tags["status"] != "failure" {
		t.Errorf("expected status failure, got %v", metric.Tags["status"])
	}
	if metric.Tags["reason"] != "compilation_error" {
		t.Errorf("expected reason compilation_error, got %v", metric.Tags["reason"])
	}
}

func TestBusinessMetrics_PipelineDuration(t *testing.T) {
	mock := &businessMockProducer{}
	ctx := context.Background()
	client := metrics.NewKafkaMetricsClient(mock, "test", ctx)
	business := metrics.NewBusinessMetrics(client)

	business.PipelineDuration("algorithms-101", 45.67)

	if len(mock.metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(mock.metrics))
	}

	metric := mock.metrics[0]
	if metric.Name != "pipeline_duration_seconds" {
		t.Errorf("expected name %q, got %q", "pipeline_duration_seconds", metric.Name)
	}
	if metric.Value != 45.67 {
		t.Errorf("expected value 45.67, got %f", metric.Value)
	}
}

func TestBusinessMetrics_MultipleCalls(t *testing.T) {
	mock := &businessMockProducer{}
	ctx := context.Background()
	client := metrics.NewKafkaMetricsClient(mock, "test", ctx)
	business := metrics.NewBusinessMetrics(client)

	business.TaskSubmitted("algorithms-101", "hw1")
	business.TaskCompleted("algorithms-101", "hw1")
	business.PipelineDuration("algorithms-101", 42.0)

	if len(mock.metrics) != 3 {
		t.Errorf("expected 3 metrics, got %d", len(mock.metrics))
	}
}
