# Kafka client (producer)

Минимальный клиент для публикации метрик в Kafka через `kafka-go`.

## Конфигурация

Клиент использует `KafkaConfig` из `internal/config`. Пример YAML:
(Здесь задаем параметры куда будем писать и как будем писать)

```yaml
kafka:
  brokers:
    - "localhost:9092"
  topic_metrics: "metrics"
  required_acks: -1
  compression: "snappy"
  allow_auto_topic_creation: true
  dial_timeout: 5s
  read_timeout: 5s
  write_timeout: 5s
  batch_timeout: 10ms
  batch_size: 100
  max_attempts: 3
  min_bytes: 1024
  max_bytes: 10485760
```

## Использование

(Также можно посмотреть пример в тестах)

```go
// Загружаем конфигурацию (брокеры, топик, acks, таймауты и т.д.)
cfg, err := config.Load("config/config.yaml")
if err != nil {
    // handle error
}

// Создаем продюсер. Он будет писать метрики в Kafka.
producer, err := kafka.NewProducer(cfg.Kafka)
if err != nil {
    // handle error
}
// Освобождаем ресурсы и закрываем соединения при завершении работы.
defer producer.Close()

// Формируем метрику: имя, значение и произвольные теги.
metric := kafka.Metric{
    Name:  "requests_total",
    Value: 1,
    Tags: map[string]string{
        "route": "/v1/echo",
        "status": "200",
    },
}

// Отправляем метрику.
// context.Background() — базовый контекст без таймаута/отмены.
// В реальном коде лучше передавать контекст запроса или таймаут:
// ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
// defer cancel()
_ = producer.PublishMetric(context.Background(), metric)
```

## Гарантии доставки

`required_acks: -1` + синхронный writer (`Async: false`) дают **at‑least‑once**.
Для полной идемпотентности нужно менять модель и включать соответствующие настройки Kafka.
(Можно еще указать required_acks = 1, тогда подтверждение будем ждать только от лидера партиции, но пока оставим так(-1)): будем ждать от всех реплик)

## Тесты

Юнит‑тесты: `go test ./internal/kafka/ut`

Функциональные: `go test -tags=integration ./internal/kafka/ft`
