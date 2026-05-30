GO ?= go
MOCKGEN := $(shell go env GOPATH)/bin/mockgen
MODULE_NAME := fcstask-backend
BINARY_NAME := fcstask-api
DOCKER_IMAGE_NAME ?= miruken/$(MODULE_NAME)-backend
DOCKER_IMAGE_TAG ?= 0.1.0
COMPOSE ?= docker compose -f docker-compose.yaml
GO_MIN_VERSION := 1.26

.PHONY: check-go check-docker init tidy migrate install-tools gen test test-integration-db postgreplication-up \
	deps db-up db-down db-wait up down stop-api run-api \
	docker-build docker-run docker-test docker-push ci-local ci

check-go:
	@command -v $(GO) >/dev/null 2>&1 || { \
		echo "ERROR: Go is not installed or not in PATH."; \
		echo "Install Go $(GO_MIN_VERSION)+: https://go.dev/dl/"; \
		echo "Example: export PATH=\$$PATH:/usr/local/go/bin"; \
		exit 1; \
	}
	@$(GO) version | grep -qE 'go$(GO_MIN_VERSION)|go1\.(2[6-9]|[3-9][0-9])' || { \
		echo "ERROR: Go $(GO_MIN_VERSION)+ required (see go.mod). Current:"; \
		$(GO) version; \
		exit 1; \
	}

check-docker:
	@command -v docker >/dev/null 2>&1 || { \
		echo "ERROR: docker is not installed or not in PATH."; \
		exit 1; \
	}
	@docker compose version >/dev/null 2>&1 || docker-compose version >/dev/null 2>&1 || { \
		echo "ERROR: docker compose plugin is not available."; \
		exit 1; \
	}
	@docker info >/dev/null 2>&1 || { \
		echo "ERROR: no access to Docker (permission denied on /var/run/docker.sock)."; \
		echo "You are in group 'docker', but this shell was started before that."; \
		echo "Run:  newgrp docker"; \
		echo "Or log out and log in again, then:  make up"; \
		exit 1; \
	}

init:
	@echo "🔧 Initializing repo: $(MODULE_NAME)..."
	@if [ ! -f go.mod ]; then \
		$(GO) mod init $(MODULE_NAME) && \
		echo "✅ go.mod created"; \
	else \
		echo "⚠️  go.mod already exists"; \
	fi

tidy:
	@echo "🧹 Tidying dependencies..."
	@$(GO) mod tidy
	@echo "✅ go.mod & go.sum updated"

# Миграции БД
migrate:
	@echo "Running database migrations..."
	$(GO) run ./migrate/main.go

# Зависимости Go и инструменты генерации (без Docker и без API)
deps: check-go init tidy install-tools
	@echo "Downloading Go module dependencies..."
	@$(GO) mod download
	@echo "Go dependencies ready"

# Поднять всё с нуля: deps → Postgres (docker) → codegen → миграции → API (foreground)
up: check-go check-docker deps db-up db-wait gen migrate run-api

# Остановить API и контейнеры Postgres
down: stop-api db-down
	@echo "Development stack stopped"

stop-api:
	@echo "Stopping API on port 8080 if running..."
	@-pkill -f "[g]o run ./internal/cmd/" 2>/dev/null || true
	@-pkill -f "[f]cstask-api" 2>/dev/null || true
	@if command -v lsof >/dev/null 2>&1; then \
		pids=$$(lsof -ti:8080 2>/dev/null); \
		[ -n "$$pids" ] && kill $$pids 2>/dev/null || true; \
	elif command -v fuser >/dev/null 2>&1; then \
		fuser -k 8080/tcp 2>/dev/null || true; \
	fi

# Локальный стек: Postgres из docker-compose
db-up: check-docker
	@echo "Starting PostgreSQL primary and replica (docker compose)..."
	@$(COMPOSE) up -d
	@echo "PostgreSQL primary: localhost:6432, replica: localhost:6433, database fcstask"

db-down:
	@echo "Stopping PostgreSQL stack..."
	@$(COMPOSE) down
	@echo "PostgreSQL stack stopped"

db-wait:
	@echo "Waiting for PostgreSQL primary to accept connections (timeout 60s)..."
	@i=0; \
	while [ $$i -lt 60 ]; do \
		if $(COMPOSE) exec -T postgres-primary pg_isready -U postgres -d fcstask 2>/dev/null; then \
			echo "PostgreSQL primary is ready"; \
			exit 0; \
		fi; \
		i=$$((i+1)); \
		sleep 1; \
	done; \
	echo "ERROR: timeout waiting for PostgreSQL primary"; \
	exit 1

run-api:
	@echo "Starting API at http://localhost:8080 (Ctrl+C to stop, or: make down)"
	@$(GO) run ./internal/cmd/

install-tools: check-go
	@echo "📦 Checking tools..."
	@test -x "$(MOCKGEN)" || $(GO) install github.com/golang/mock/mockgen@latest
	@echo "✅ Tools ready"

gen: check-go
	@echo "Generating API code from OpenAPI..."
	$(GO) tool oapi-codegen -generate types,skip-prune -package api -o internal/api/types.gen.go api/openapi.yaml
	$(GO) tool oapi-codegen -generate server -package api -o internal/api/server.gen.go api/openapi.yaml
	@echo "✅ Code generation completed"

test: gen
	@echo "🧪 Running tests..."
	@$(GO) test ./... -v
	@echo "✅ Tests completed"

postgreplication-up: db-up
	@echo "PostgreSQL replication stack is up (alias for db-up)"

test-integration-db:
	$(GO) test -tags=integration -count=1 -v ./internal/db/...
