GO ?= go
GOPATH ?= $(HOME)/go
export PATH := $(PATH):$(GOPATH)/bin
OAPI_CODEGEN := $(GOPATH)/bin/oapi-codegen
MOCKGEN := $(GOPATH)/bin/mockgen
MODULE_NAME := fcstask-backend
BINARY_NAME := fcstask-api
DOCKER_IMAGE_NAME ?= miruken/$(MODULE_NAME)-backend
DOCKER_IMAGE_TAG ?= 0.1.0
COMPOSE ?= docker compose -f docker-compose.yaml
GO_MIN_VERSION := 1.26

.PHONY: check-go check-docker init tidy migrate install-tools gen test test-integration-db postgreplication-up \
	deps db-up db-down db-wait up down stop-api run-api \
	monitoring-up monitoring-down monitoring-logs load-test \
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
	@echo "📦 Installing tools..."
	@test -x "$(OAPI_CODEGEN)" || $(GO) install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	@test -x "$(MOCKGEN)" || $(GO) install github.com/golang/mock/mockgen@latest
	@echo "✅ Tools installed"

gen: install-tools
	@echo "Generating API code from OpenAPI..."
	@if test -x "$(OAPI_CODEGEN)"; then \
		echo "oapi-codegen is already installed"; \
	else \
		echo "Installing oapi-codegen..."; \
		$(GO) install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest; \
	fi
	@echo "Generating types..."
	$(OAPI_CODEGEN) -generate types,skip-prune -package api -o internal/api/types.gen.go api/openapi.yaml
	@echo "Generating server..."
	$(OAPI_CODEGEN) -generate server -package api -o internal/api/server.gen.go api/openapi.yaml
	@echo "Code generation completed!"
	@echo "🔄 Generating code..."
	@go generate ./...
	@echo "✅ Code generation completed"

test: gen
	@echo "🧪 Running tests..."
	@$(GO) test ./... -v
	@echo "✅ Tests completed"

postgreplication-up: db-up
	@echo "PostgreSQL replication stack is up (alias for db-up)"

test-integration-db:
	$(GO) test -tags=integration -count=1 -v ./internal/db/...

COMPOSE_MONITORING ?= docker compose -f docker-compose.yaml --profile monitoring

monitoring-up: check-docker
	@echo "Starting monitoring stack (prometheus, alertmanager, exporters)..."
	@$(COMPOSE_MONITORING) up -d
	@echo "Stack is up. UIs:"
	@echo "  Prometheus:        http://localhost:9090"
	@echo "  Alertmanager:      http://localhost:9093"
	@echo "  Postgres exporter: http://localhost:9187/metrics"
	@echo "  SQL exporter:      http://localhost:9399/metrics"
	@echo "Make sure the app is running on host:8081/metrics (run: make run-api)"

monitoring-down:
	@echo "Stopping monitoring stack..."
	@$(COMPOSE_MONITORING) stop prometheus alertmanager postgres-exporter sql-exporter
	@$(COMPOSE_MONITORING) rm -f prometheus alertmanager postgres-exporter sql-exporter
	@echo "Monitoring stack stopped (Postgres primary/replica left running)"

monitoring-logs:
	@$(COMPOSE_MONITORING) logs -f prometheus alertmanager postgres-exporter sql-exporter

docker-build:
	@echo "🐳 Building Docker image..."
	@docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .
	@echo "✅ Docker image built"

docker-run: docker-build
	@echo "🚀 Running container on http://localhost:8080"
	@docker run --rm -p 8080:8080 $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

docker-test:
	@echo "🧪 Running tests inside container..."
	docker run --rm \
		-v "$(PWD):/app" \
		-w /app \
		golang:1.25-alpine \
		go test ./... -v

docker-push:
	@if [ -z "$$CI" ] && [ -z "$$FORCE_PUSH" ]; then \
		echo "🛑 ERROR: Refusing to push from local machine."; \
		echo "💡 Run with FORCE_PUSH=1 to override (not recommended)."; \
		exit 1; \
	fi
	@echo "📤 Pushing image to registry..."
	@docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
	@echo "✅ Pushed: $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)"

ci-local: init tidy test docker-build
	@echo "✅ Local CI check passed!"

ci: ci-local docker-push
	@echo "✅ Full CI pipeline completed!"
