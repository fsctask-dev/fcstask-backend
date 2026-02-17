# Система сбора метрик

## Архитектура
FcsTask → Kafka → Telegraf → Prometheus → Grafana

## Компоненты

- **metrics** — пакет для отправки метрик из кода
- **middleware** — автоматический сбор HTTP метрик
- **Telegraf** — читает метрики из Kafka, отдаёт в Prometheus (см. fcstask/config/telegraf.conf)

# Пакет metrics

```text
internal/metrics/

metrics/
├── client.go     # Клиент для отправки метрик в Kafka
├── business.go   # Бизнес-метрики (задачи, пайплайны)
└── interface.go  # Общие интерфейсы (Producer, MetricsClient)
```
## Интерфейс MetricsClient
```go
type MetricsClient interface {
// Inc - увеличивает счётчик на 1
Inc(name string, tags map[string]string)

    // Add - увеличивает счётчик на произвольное значение
    Add(name string, value float64, tags map[string]string)
    
    // Gauge - устанавливает текущее значение метрики
    Gauge(name string, value float64, tags map[string]string)
    
    // Histogram - записывает наблюдение (время, размер и т.п.)
    Histogram(name string, value float64, tags map[string]string)
    
    // Close - закрывает клиент
    Close() error
}
```
Где использовать: везде, где нужно отправлять метрики.

## KafkaMetricsClient
Реализация MetricsClient, отправляющая метрики в Kafka.

```go
// Создание клиента
client := metrics.NewKafkaMetricsClient(producer, "edu.metrics", ctx)

// Отправка метрик
client.Inc("http_requests_total", map[string]string{"path": "/api"})
client.Histogram("response_time", 0.042, map[string]string{"method": "GET"})
client.Gauge("active_connections", 42, nil)
```
Требования:

producer должен реализовывать интерфейс Producer (метод PublishMetric)

Kafka топик должен существовать или AllowAutoTopicCreation: true

## BusinessMetrics
Готовые бизнес-метрики для системы проверки задач.

```go
business := metrics.NewBusinessMetrics(client)

// Задача отправлена на проверку
business.TaskSubmitted("algorithms-101", "hw1")

// Задача успешно проверена
business.TaskCompleted("algorithms-101", "hw1", "success")

// Задача упала с ошибкой
business.TaskFailed("algorithms-101", "hw1", "timeout")

// Время выполнения пайплайна
business.PipelineDuration("algorithms-101", 45.67)
```

## Метрики:

| Метод            | Имя метрики                | Теги                               |
|------------------|----------------------------|------------------------------------|
| TaskSubmitted    | tasks_submitted_total      | course, assignment                |
| TaskCompleted    | tasks_completed_total      | course, assignment, status        |
| TaskFailed       | tasks_failed_total         | course, assignment, reason        |
| PipelineDuration | pipeline_duration_seconds  | course                            |
