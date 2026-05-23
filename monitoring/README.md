
# Monitoring

## Быстрый старт

#### 1. Настроить переменные окружения

```bash
cp monitoring/.env.example monitoring/.env
cp monitoring/bot/.env.example monitoring/bot/.env
```

Заполнить `monitoring/.env`:

| Переменная        | Описание                                     |
| ----------------- | -------------------------------------------- |
| `BOT_DB_PASSWORD` | Пароль PostgreSQL для бота                   |
| `ACTIVE_RULE`     | PromQL-выражение для алерта (debug или prod) |

Заполнить `monitoring/bot/.env`:

| Переменная  | Описание                 |
| ----------- | ------------------------ |
| `BOT_TOKEN` | Токен бота от @BotFather |

### 2. Запустить

```bash
make monitoring-up
```

### 3. Проверить

```
Prometheus   → http://localhost:9090
Alertmanager → http://localhost:9093
Bot          → http://localhost:9094
Telegram     → https://t.me/fcstask_monitor_bot
```

## Алерт-правила

В `monitoring/.env` два варианта `ACTIVE_RULE`:

```bash
# Debug: срабатывает на любой 500 за последнюю минуту
ACTIVE_RULE=sum(rate(echo_requests_total{code="500"}[1m]))>0

# Prod: срабатывает если >10% запросов с ошибкой за 5 минут
#ACTIVE_RULE=sum(rate(echo_requests_total{code="500"}[5m])) / sum(rate(echo_requests_total[5m]))>0.1
```

Переключение: раскомментировать нужную строку и перезапустить `make monitoring-up`.

## Telegram-бот

Бот: [@fcstask_monitor_bot](https://t.me/fcstask_monitor_bot)

## Makefile команды

```bash
make monitoring-up      # Запустить стек
make monitoring-down    # Остановить стек
```

## Ручное тестирование алерта

```bash
curl -X POST http://localhost:9094/alert \
  -H "Content-Type: application/json" \
  -d '{
    "status": "firing",
    "alerts": [{
      "status": "firing",
      "labels": {"alertname": "HighHTTP500Errors", "severity": "critical"},
      "annotations": {"summary": "Test alert"},
      "startsAt": "2026-01-01T00:00:00Z",
      "endsAt": "0001-01-01T00:00:00Z"
    }]
  }'
```
