# fcstask

## Quick start

Нужны **Go**, **Docker** и **Make**.

```bash
make up    # зависимости, Postgres (master :6432, replica :6433), миграции, API
make down  # остановить API и контейнеры БД
```

API: [http://localhost:8080](http://localhost:8080). Конфиг: `config/config.yaml`

| Команда             | Описание |
|---------------------|----------|
| `make up`           | Поднять локальный стек с нуля (см. Quick start). |
| `make down`         | Остановить API и Postgres. |
| `make init`         | Инициализирует Go-модуль (`go.mod`), если он ещё не создан. |
| `make tidy`         | Обновляет зависимости (`go mod tidy`). |
| `make install-tools`| Устанавливает необходимые инструменты: `oapi-codegen` и `mockgen`. |
| `make gen`          | Запускает генерацию кода (например, из OpenAPI-спецификации). |
| `make build`        | Собирает бинарный файл `fcstask-api` из `internal/cmd/main.go`. |
| `make test`         | Запускает все unit-тесты с подробным выводом (`go test ./... -v`). |
| `make docker-build` | Собирает Docker-образ с тегом `miruken/fcstask-backend:0.1.0`. |
| `make docker-run`   | Собирает образ и запускает контейнер на `http://localhost:8080`. |
| `make docker-test`  | Запускает тесты внутри временного контейнера на базе `golang:1.25-alpine`. |
| `make ci`           | То же, что `ci-local`, плюс пушит образ в Docker Registry (**только в CI!**). |


### 🐳 Запуск в Docker

```bash
make docker-run
```

После этого API будет доступно по адресу: [http://localhost:8080](http://localhost:8080)
