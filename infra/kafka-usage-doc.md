# Apache Kafka usage

## 1. Архитектурная схема на Go

```text
GitLab → Go Handler → Kafka → Go Services
   ↑      ↓           ↓         │
   └──────┘           │         ├───▶ DB
               Monitoring ◄─────┤
                     ↓          ├───▶ Cache  
               Analytics ◄──────┘
                              Storage
```

## 2. Producer/Writer интерфейсы на Go

### Основные интерфейсы

```go
type EventProducer interface {
    ProduceGitLabEvent(ctx context.Context, event GitLabEvent) error
    ProducePipelineStatus(ctx context.Context, status PipelineStatus) error
    ProduceTestResult(ctx context.Context, result TestResult) error
    Close() error
}

type KafkaConfig struct {
    Brokers      []string          // ["kafka1:9092", "kafka2:9092", "kafka3:9092"]
    TopicPrefix  string            // "edu."
    Compression  string            // "snappy"
    RequiredAcks int               // -1 (WaitForAll)
}
```

## 3. Consumer/Reader интерфейсы на Go
```go
type ConsumerGroup interface {
    Consume(ctx context.Context, handler MessageHandler) error
    Commit() error
    Close() error
}

// Пример Consumer Groups:
groups := map[string]ConsumerConfig{
    "pipeline-orchestrators": {
        Topics:    []string{"gitlab.push"},
        Instances: 3,  // Пример, можно увеличить в зависимости от нагрузки
        Handler:   startPipelineHandler,
    },
    "result-processors": {
        Topics:    []string{"pipeline.results"},
        Instances: 2,
        Handler:   processResultsHandler,
    },
    "notification-senders": {
        Topics:    []string{"grading.completed"},
        Instances: 1,
        Handler:   sendNotificationHandler,
    },
}
```

## 4. Используемая очередь - Apache Kafka

Рассматривались варианты: RabbitMQ, NATS (с JetStream), Apache Kafka

### 4.1 RabbitMQ
#### Плюсы:

* Простой в понимании
* Есть веб-интерфейс, где видно все очереди
* Хорошая документация

#### Минусы:

* Сложность обеспечения строгого порядка сообщений
* Ограничения горизонтального масштабирования
* Высокая нагрузка на брокер при большом числе consumers

### 4.2 NATS с JetStream
#### Плюсы:

* Очень быстрый
* Простой в настройке

#### Минусы:

* Меньше production-опыта в аналогичных сценариях
* Ограниченные инструменты мониторинга
* Требует доработки под специфические требования

### 4.3 Apache Kafka
#### Преимущества для нашего сценария:

* Высокая производительность
* Гарантия доставки
* Партиционирование по курсам для параллельной обработки
* Сохранение порядка сообщений в пределах партиции
* Возможность replay событий для отладки
* Широкая экосистема инструментов

### Итого: выбираем Apache Kafka

## 5. Вычитка

### Consumer Groups
Основная вычитка через них (см 3 пункт)

#### Как работает:

* Каждая группа независимо читает сообщения
* Внутри группы партиции распределяются между инстансами
* Каждое сообщение обрабатывается одним consumer в группе

### Партиционирование

#### Принцип
```go
partition = hash(key) % total_partitions
```
(как вариант хэша "course:assignment:student")

### Стратегии чтения
* Batch processing
* Stream Processing
* Exactly-once processing

### Поток данных
```text
Сообщение в Kafka
    ↓
Consumer.poll() → получение сообщения
    ↓
Десериализация (Avro/JSON)
    ↓
Валидация схемы
    ↓
Обработка бизнес-логикой
    ↓
Сохранение результата
    ↓
Commit offset (подтверждение обработки)
    ↓
Следующее сообщение
```

## 6. Итог
Apache Kafka с Go — оптимальный выбор для системы проверки задач:
* Высокая производительность — обработка пиковых нагрузок контрольных
* Гарантии доставки — ни одна задача не потеряется
* Масштабируемость — легко добавлять обработчики под нагрузку
* Мониторинг — полная observability через Prometheus
* Отказоустойчивость — репликация на 3 нодах

Используем Sarama (production-ready) или kafka-go (макс. производительность)
