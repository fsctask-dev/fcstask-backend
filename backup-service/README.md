# backup-service

Микросервис резервного копирования PostgreSQL с поддержкой PITR и архивации WAL.

## Запуск

Все команды принимают флаг `--config FILE` (по умолчанию `config.yaml`).

```bash
backup-service <команда> [--config config.yaml] [флаги]
```

### Команды

| Команда       | Назначение                                                        |
|---------------|-------------------------------------------------------------------|
| `serve`       | Демон: создаёт бэкапы по расписанию (cron) + health-эндпоинт      |
| `backup`      | Разовый бэкап (полный или инкрементальный)                        |
| `restore`     | Восстановление в отдельную БД на момент времени или на последнюю |
| `list`        | Список доступных бэкапов                                          |
| `verify`      | Проверка целостности бэкапов по контрольным суммам                |
| `wal-backup`  | Базовый бэкап для WAL-режима                                      |
| `wal-archive` | Демон архивации WAL-сегментов                                     |
| `wal-restore` | Восстановление из WAL на момент времени                          |
| `wal-list`    | Список базовых бэкапов WAL                                        |

## Сценарии использования

### 1. Бэкапы по расписанию (основной режим)

Запуск демона. Расписание задаётся в `config.yaml` → `cron.schedule` (по умолчанию каждый день в 02:00). Каждый N-й бэкап полный, между ними инкрементальные

```bash
backup-service serve --config config.yaml
```

Состояние доступно по health-эндпоинту (`health.addr`, по умолчанию `:8080`):

```bash
curl http://localhost:8080/healthz   # живость
curl http://localhost:8080/readyz    # готовность
```

### 2. Разовый бэкап вручную

```bash
backup-service backup --config config.yaml               # авто (полный/инкрементальный)
backup-service backup --config config.yaml --full        # принудительно полный
backup-service backup --config config.yaml --incremental # принудительно инкрементальный
```

### 3. Просмотр и проверка бэкапов

```bash
backup-service list   --config config.yaml   # таблица всех бэкапов
backup-service verify --config config.yaml   # проверка контрольных сумм
```

### 4. Восстановление (PITR)

Восстановление выполняется в **отдельную** БД (`restore_target`), не в источник

```bash
# на конкретный момент времени (RFC3339, UTC)
backup-service restore --config config.yaml --time 2026-06-07T00:00:00Z

# на самое последнее доступное состояние
backup-service restore --config config.yaml --latest
```

### 5. Непрерывная архивация WAL

Требует `wal.enabled: true` в конфиге

```bash
# 1. снять базовый бэкап
backup-service wal-backup --config config.yaml

# 2. запустить демон архивации WAL-сегментов
backup-service wal-archive --config config.yaml

# 3. посмотреть список базовых бэкапов
backup-service wal-list --config config.yaml

# 4. восстановить из WAL на момент времени или на последний
backup-service wal-restore --config config.yaml --time 2026-06-07T00:00:00Z
backup-service wal-restore --config config.yaml --latest
```

## Запуск в Docker

Образ собирается на базе `postgres-alpine`

```bash
docker build -t backup-service -f backup-service/Dockerfile .
docker run -v /data/backups:/var/backups/postgres backup-service
```

## Примечания

- Одновременно выполняется только одна операция — защита через lock-файл в каталоге бэкапов
- Время для `--time` указывается в формате RFC3339 (например `2026-06-07T00:00:00Z`)
- Параметры подключения к БД, расписание, хранение и ретенция настраиваются в `config.yaml`
