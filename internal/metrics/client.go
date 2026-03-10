package metrics

import (
	"context"
	"time"

	"fcstask/internal/kafka"
)

type Producer interface {
	PublishMetric(ctx context.Context, metric kafka.Metric) error
	Close() error
}

type MetricsClient interface {
	Inc(name string, tags map[string]string)

	Add(name string, value float64, tags map[string]string)

	Gauge(name string, value float64, tags map[string]string)

	Histogram(name string, value float64, tags map[string]string)

	Close() error
}

type KafkaMetricsClient struct {
	producer Producer
	topic    string
	ctx      context.Context
}

func NewKafkaMetricsClient(producer Producer, topic string, ctx context.Context) *KafkaMetricsClient {
	return &KafkaMetricsClient{
		producer: producer,
		topic:    topic,
		ctx:      ctx,
	}
}

func (mc *KafkaMetricsClient) Inc(name string, tags map[string]string) {
	mc.Add(name, 1, tags)
}

func (mc *KafkaMetricsClient) Add(name string, value float64, tags map[string]string) {
	_ = mc.producer.PublishMetric(mc.ctx, kafka.NewMetric(name, value, tags, time.Now()))
}

func (mc *KafkaMetricsClient) Gauge(name string, value float64, tags map[string]string) {
	_ = mc.producer.PublishMetric(mc.ctx, kafka.NewMetric(name, value, tags, time.Now()))
}

func (mc *KafkaMetricsClient) Histogram(name string, value float64, tags map[string]string) {
	_ = mc.producer.PublishMetric(mc.ctx, kafka.NewMetric(name, value, tags, time.Now()))
}

func (mc *KafkaMetricsClient) Close() error {
	return mc.producer.Close()
}
